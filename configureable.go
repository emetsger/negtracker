package main

type Configurable interface {
	Configure(config interface{})
}
