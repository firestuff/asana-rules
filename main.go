package main

import "fmt"

import "github.com/firestuff/asana-rules/asana"

func main() {
  a := asana.NewClientFromEnv()

  /*
  me, err := a.GetMe()
  if err != nil {
    panic(err)
  }
  */

  wrk, err := a.GetWorkspace()
  if err != nil {
    panic(err)
  }

  prjs, err := a.GetProjects(wrk.GID)
  if err != nil {
    panic(err)
  }

  for _, prj := range prjs {
    fmt.Printf("%#v\n", prj)
  }
}
