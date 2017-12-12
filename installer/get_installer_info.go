package installer

import (
	"io"
	"path/filepath"

	"github.com/go-errors/errors"
	"github.com/itchio/butler/archive"
	"github.com/itchio/butler/archive/backends/bah"
	"github.com/itchio/butler/archive/backends/szah"
	"github.com/itchio/butler/configurator"
	"github.com/itchio/wharf/eos"
	"github.com/itchio/wharf/state"
)

func GetInstallerInfo(consumer *state.Consumer, file eos.File) (*InstallerInfo, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	target := stat.Name()
	ext := filepath.Ext(target)
	name := filepath.Base(target)

	if typ, ok := installerForExt[ext]; ok {
		consumer.Infof("%s: choosing installer '%s'", name, typ)
		var archiveHandlerName string
		if typ == InstallerTypeArchive {
			switch ext {
			case ".zip":
				archiveHandlerName = (&bah.Handler{}).Name()
			default:
				archiveHandlerName = (&szah.Handler{}).Name()
			}
		}
		return &InstallerInfo{
			Type:               typ,
			ArchiveHandlerName: archiveHandlerName,
		}, nil
	}

	consumer.Infof("%s: probing as archive", name)
	handler, err := archive.TryOpen(&archive.TryOpenParams{
		File:     file,
		Consumer: consumer,
	})
	if err == nil {
		consumer.Infof("%s: is archive", name)
		return &InstallerInfo{
			Type:               InstallerTypeArchive,
			ArchiveHandlerName: handler.Name(),
		}, nil
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	consumer.Infof("%s: probing with configurator", name)

	candidate, err := configurator.Sniff(file, target, stat.Size())
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	_, err = file.Seek(0, io.SeekStart)
	if err != nil {
		return nil, errors.Wrap(err, 0)
	}

	if candidate != nil {
		typ := getInstallerTypeForCandidate(consumer, name, candidate)
		if typ != InstallerTypeUnknown {
			return &InstallerInfo{Type: typ}, nil
		}
	} else {
		consumer.Infof("%s: nil candidate, configurator has forsaken us")
	}

	return &InstallerInfo{
		Type: InstallerTypeUnknown,
	}, nil
}

func getInstallerTypeForCandidate(consumer *state.Consumer, name string, candidate *configurator.Candidate) InstallerType {
	switch candidate.Flavor {

	case configurator.FlavorNativeWindows:
		if candidate.WindowsInfo != nil && candidate.WindowsInfo.InstallerType != "" {
			typ := (InstallerType)(candidate.WindowsInfo.InstallerType)
			consumer.Infof("%s: windows installer of type %s", name, typ)
			return typ
		}

		consumer.Infof("%s: native windows executable, but not an installer", name)
		return InstallerTypeNaked

	case configurator.FlavorNativeMacos:
		consumer.Infof("%s: native macOS executable", name)
		return InstallerTypeNaked

	case configurator.FlavorNativeLinux:
		consumer.Infof("%s: native linux executable", name)
		return InstallerTypeNaked

	case configurator.FlavorScript:
		consumer.Infof("%s: script", name)
		if candidate.ScriptInfo != nil && candidate.ScriptInfo.Interpreter != "" {
			consumer.Infof("...with interpreter %s", candidate.ScriptInfo.Interpreter)
		}
		return InstallerTypeNaked

	case configurator.FlavorScriptWindows:
		consumer.Infof("%s: windows script", name)
		return InstallerTypeNaked

	}

	return InstallerTypeUnknown
}