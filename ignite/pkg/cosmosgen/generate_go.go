package cosmosgen

import (
	"os"
	"path/filepath"

	"github.com/otiai10/copy"
	"github.com/pkg/errors"
)

const goTemplate = "buf.gen.gogo.yaml"

func (g *generator) generateGo() error {
	// create a temporary dir to locate generated code under which later only some of them will be moved to the
	// app's source code. this also prevents having leftover files in the app's source code or its parent dir - when
	// command executed directly there - in case of an interrupt.
	tmp, err := os.MkdirTemp("", "")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmp)

	protoPath := filepath.Join(g.appPath, g.protoDir)

	// code generate for each module.
	if err := g.buf.Generate(
		g.ctx,
		protoPath,
		tmp,
		goTemplate,
	); err != nil {
		return err
	}

	// move generated code for the app under the relative locations in its source code.
	generatedPath := filepath.Join(tmp, g.o.gomodPath)

	_, err = os.Stat(generatedPath)
	if err == nil {
		err = copy.Copy(generatedPath, g.appPath)
		if err != nil {
			return errors.Wrap(err, "cannot copy path")
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	return nil
}
