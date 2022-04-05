package protoanalysis

import (
	"context"
	"os"
	"path/filepath"

	"github.com/emicklei/proto"
	"github.com/ignite-hq/cli/starport/pkg/localfs"
	"github.com/pkg/errors"
)

const optionGoPkg = "go_package"

// parser parses proto packages.
type parser struct {
	packages []*pkg
}

// parse parses proto files in the fs that matches with pattern and returns
// the low level representations of proto packages.
func parse(ctx context.Context, path, pattern string) ([]*pkg, error) {
	pr := &parser{}

	paths, err := localfs.Search(path, pattern)
	if err != nil {
		return nil, err
	}

	for _, path := range paths {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if err := pr.parseFile(path); err != nil {
			return nil, errors.Wrapf(err, "file: %s", path)
		}
	}

	return pr.packages, nil
}

// pkg represents a proto package.
type pkg struct {
	// name of the proto package.
	name string

	// directory of the proto package in the fs.
	dir string

	// files is a list of proto files that construct a proto package.
	files []file
}

// file represents a parsed proto file.
type file struct {
	// path of the proto file in the fs.
	path string

	// parsed data.
	pkg      *proto.Package
	imports  []string // imported protos.
	options  []*proto.Option
	messages []*proto.Message
	services []*proto.Service
}

func (p *pkg) options() (o []*proto.Option) {
	for _, f := range p.files {
		o = append(o, f.options...)
	}

	return
}

func (p *pkg) messages() (m []*proto.Message) {
	for _, f := range p.files {
		m = append(m, f.messages...)
	}

	return
}

func (p *pkg) services() (s []*proto.Service) {
	for _, f := range p.files {
		s = append(s, f.services...)
	}

	return
}

func (p *parser) parseFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	def, err := proto.NewParser(f).Parse()
	if err != nil {
		return err
	}

	var pkgName string

	proto.Walk(
		def,
		proto.WithPackage(func(p *proto.Package) { pkgName = p.Name }),
	)

	var pp *pkg
	for _, v := range p.packages {
		if pkgName == v.name {
			pp = v
			break
		}
	}
	if pp == nil {
		pp = &pkg{
			name: pkgName,
			dir:  filepath.Dir(path),
		}
		p.packages = append(p.packages, pp)
	}

	pf := file{
		path: path,
	}

	proto.Walk(
		def,
		proto.WithPackage(func(p *proto.Package) { pf.pkg = p }),
		proto.WithImport(func(s *proto.Import) { pf.imports = append(pf.imports, s.Filename) }),
		proto.WithOption(func(o *proto.Option) { pf.options = append(pf.options, o) }),
		proto.WithMessage(func(m *proto.Message) { pf.messages = append(pf.messages, m) }),
		proto.WithService(func(s *proto.Service) { pf.services = append(pf.services, s) }),
	)

	pp.files = append(pp.files, pf)

	return nil
}
