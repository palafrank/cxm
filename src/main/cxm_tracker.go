package main
import (
  "os"
  "fmt"
  "pktlib"
)

type pktFlow struct {
  signature []byte
  scid int
  port int
}

var flows []*pktFlow
var flowChannel chan pktFlow

func PktTrackerInit() {
  flowChannel = make(chan pktFlow)
  go PktTrackerDump(flowChannel)
}

func PktTracker(msg []byte, scid int, port int) {
  flow := new(pktFlow)
  signature := pktutil.GetPktSignature(pktSigOffset, pktSigLen, msg)

  flow.signature = signature
  flow.scid = scid
  flow.port = port
  flows = append(flows, flow)
  if(len(flowChannel) < cap(flowChannel)) {
    flow, flows = flows[len(flows)-1], flows[:len(flows)-1]
    flowChannel <- *flow
  }
}

func PktTrackerDump(ch chan pktFlow) {
  f, err := os.OpenFile(trackFile, os.O_RDWR|os.O_CREATE, 0755)
  if(err != nil) {
    fmt.Println("Error opening tracker file")
    return
  }
  fmt.Println("Successfully started packet tracker: ", trackFile)
  for {
    select {
    case data := <- ch:
      for ind := range data.signature {
        f.Write([]byte(fmt.Sprintf("%02x", data.signature[ind])))
      }
      //f.Write(data.signature)
      mystring := fmt.Sprintf("::SCID:%d::PORT:%d\n", data.scid, data.port)
      f.Write([]byte(mystring))
      f.Sync()
    }
  }
}
