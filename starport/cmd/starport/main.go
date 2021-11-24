package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	starportcmd "github.com/tendermint/starport/starport/cmd"
	"github.com/tendermint/starport/starport/pkg/clictx"
	"github.com/tendermint/starport/starport/pkg/validation"
)

func main() {
	ctx := clictx.From(context.Background())

	// Check if this actually preruns, idk if it is right now
	err := starportcmd.New(ctx).ExecuteContext(ctx)
	if ctx.Err() == context.Canceled || err == context.Canceled {
		fmt.Println("aborted")
		return
	}

	if err != nil {
		var validationErr validation.Error

		if errors.As(err, &validationErr) {
			fmt.Println(validationErr.ValidationInfo())
		} else {
			fmt.Println(err)
		}

		os.Exit(1)
	}
}
