package main
import (
  "encoding/json"
  "os"
  "github.com/trhodeos/spicy/lib"
)

func main() {
  var spec, err = spicy.ParseSpec(os.Stdin)
  if err != nil { panic(err) }

  json.NewEncoder(os.Stdout).Encode(spec)
}
