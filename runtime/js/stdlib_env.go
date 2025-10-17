package js

import (
	"context"
	"os"
	"strings"

	"github.com/sentrie-sh/sentrie/pack"
)

func (ar *AliasRuntime) setupEnvStdLib(_ context.Context, pack *pack.PackFile) error {
	env := ar.VM.NewObject()

	// range through the os.Environ() and add to env
	for _, kv := range os.Environ() {
		key, value, found := strings.Cut(kv, "=")
		if !found {
			continue // skip malformed environment variables
		}

		if pack.Permissions.CheckEnvAccess(key) {
			if err := env.Set(key, value); err != nil {
				return err
			}
		}
	}

	if err := ar.VM.Set("env", env); err != nil {
		return err
	}
	return nil
}
