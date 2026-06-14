package cli

import (
	"errors"

	"github.com/tamnd/linkedin-cli/linkedin"
)

func isBlocked(err error) bool {
	return errors.Is(err, linkedin.ErrBlocked) || errors.Is(err, linkedin.ErrRateLimited)
}

func isNotFound(err error) bool {
	return errors.Is(err, linkedin.ErrNotFound)
}
