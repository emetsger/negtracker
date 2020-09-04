// build +integration
package store

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
)

const EnvDbKeepResults = "DB_KEEP_RESULTS"

func GenTestDbName(env, random string) string {
	return fmt.Sprintf("%s_%s_%s", os.Getenv(env), "mongostore_test", random)
}

func GenInt() string {
	return strconv.Itoa(rand.New(rand.NewSource(time.Now().UnixNano())).Intn(1000))
}
