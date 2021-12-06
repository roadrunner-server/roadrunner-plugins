package config

import (
	"fmt"

	"github.com/hashicorp/go-version"
	"github.com/spf13/viper"
	"github.com/spiral/errors"
	"github.com/spiral/roadrunner-plugins/v2/config/configs/v2_6"
	"github.com/spiral/roadrunner-plugins/v2/config/configs/v2_7"
)

const (
	v26 string = "2.6.0"
	v27 string = "2.7.0"
)

// changed sections
const (
	jobs string = "jobs"
)

// all configuration version
var versionsTable = map[string]interface{}{
	v26: &v2_6.Config{},
	v27: &v2_7.Config{},
}

// transition used to upgrade configuration.
// configuration can be upgraded between related versions (2.5 -> 2.6, but not 2.5 -> 2.7).
func transition(from, to string, v *viper.Viper) error {
	vfrom, err := version.NewSemver(from)
	if err != nil {
		return err
	}

	vto, err := version.NewSemver(to)
	if err != nil {
		return err
	}

	if !vto.GreaterThan(vfrom) {
		return errors.Str("the result version should be greater than original version")
	}

	segFrom := vfrom.Segments64()
	segTo := vto.Segments64()

	if (len(segTo) < 3 || len(segFrom) < 3) || (segTo[1]-segFrom[1]) != 1 {
		return errors.Errorf("incompatible versions passed: from: %s, to: %s", vfrom.String(), vto.String())
	}

	// use only 2 digits
	trFrom := fmt.Sprintf("%d.%d.0", segFrom[0], segFrom[1])
	trTo := fmt.Sprintf("%d.%d.0", segTo[0], segTo[1])

	switch trFrom {
	case v26:
		switch trTo { //nolint:gocritic
		case v27:
			// transition configuration from v2.6 to v2.7
			v26to27(versionsTable[v26].(*v2_6.Config), versionsTable[v27].(*v2_7.Config), v)
		}
	case v27:
		return nil
	}

	return nil
}

func v26to27(from *v2_6.Config, to *v2_7.Config, v *viper.Viper) {
	err := v.UnmarshalKey(jobs, &from.Jobs)
	if err != nil {
		panic(err)
	}
}
