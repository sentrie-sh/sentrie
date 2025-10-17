package js

import (
	"context"

	"github.com/sentrie-sh/sentrie/pack"
)

// SetupStdLib installs the standard library into the VM.
// This MUST be called before Require is used or the VM is invoked.
func (ar *AliasRuntime) SetupStdLib(ctx context.Context, pack *pack.PackFile) error {
	if err := ar.setupEnvStdLib(ctx, pack); err != nil {
		return err
	}
	return nil
}
