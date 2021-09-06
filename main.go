package main

import "fmt"

import "github.com/firestuff/asana-rules/asana"

func main() {
  a := asana.NewClientFromEnv()

  /*
  me, err := a.Me()
  if err != nil {
    panic(err)
  }
  */

  wrk, err := a.Workspace()
  if err != nil {
    panic(err)
  }

  prjs, err := a.Projects(wrk.GID)
  if err != nil {
    panic(err)
  }

  for _, prj := range prjs {
    fmt.Printf("%#v\n", prj)
  }
}
