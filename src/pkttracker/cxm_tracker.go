package pkttracker

type pktTrack struct {
	signature []byte
	src_scid  int
	src_port  int
	dst_scid  int
	dst_port  int
}

func PktTrackerInit() interface{} {
	flowChannel := make(chan *pktTrack, 100)
	go PktTrackerDump(flowChannel)
	return flowChannel
}

func PktTracker(handle interface{}, signature []byte, src_scid int, src_port int, dst_scid int, dst_port int) {
	flowChannel := handle.(chan *pktTrack)
	flow := new(pktTrack)

	flow.signature = signature
	flow.src_scid = src_scid
	flow.src_port = src_port
	flow.dst_scid = dst_scid
	flow.dst_port = dst_port
	flowChannel <- flow

}

func PktTrackerDump(ch chan *pktTrack) {
	/*
		f, err := os.OpenFile(trackFile, os.O_RDWR|os.O_CREATE, 0755)
		if err != nil {
			fmt.Println("Error opening tracker file")
			return
		}
		fmt.Println("Successfully started packet tracker: ", trackFile)
		for {
			select {
			case data := <-ch:
				for ind := range data.signature {
					f.Write([]byte(fmt.Sprintf("%02x", data.signature[ind])))
				}
				//f.Write(data.signature)
				mystring := fmt.Sprintf("::SCID:%d::PORT:%d\n", data.scid, data.port)
				f.Write([]byte(mystring))
				f.Sync()
			}
		}
	*/
	var flows []*pktTrack
	for {
		select {
		case data := <-ch:
			flows = append(flows, data)
		}
	}
}
