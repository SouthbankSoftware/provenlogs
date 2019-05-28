/*
 * @Author: guiguan
 * @Date:   2019-05-22T14:59:33+10:00
 * @Last modified by:   guiguan
 * @Last modified time: 2019-05-25T10:21:11+10:00
 */

package main

import (
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	counter := 0

	for {
		r := rand.Intn(1990) + 10

		logger.Info(fmt.Sprintf("test %v", counter), zap.Int("delay", r))

		counter++
		time.Sleep(time.Duration(r) * time.Millisecond)
	}

}
