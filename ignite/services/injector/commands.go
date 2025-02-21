package injector

import "context"

type Command struct {
}

func (i *injector) AddCommand(ctx context.Context, c *Command) error {
	panic("unimplemented")
}
