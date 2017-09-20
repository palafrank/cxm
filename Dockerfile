FROM adm_switch
CMD cd adm_switch && go run cxm.go cxm_comp.go cxm_conn.go cxm_parser.go cxm_pkt.go link_pkt.go
