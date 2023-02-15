package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	ysrl "github.com/azyablov/ysrl/srl/srlv22m6p4"
	"github.com/openconfig/gnmic/api"
	"github.com/openconfig/ygot/ygot"
	"google.golang.org/protobuf/encoding/prototext"
)

func main() {
	// Target definition
	// TODO: add TLS certs
	t, err := api.NewTarget(api.Name("clab-2nd-srl1"),
		api.Address("clab-2nd-srl1:57400"),
		api.Username("admin"),
		api.Password("admin"),
		api.Timeout(time.Second*5),
		api.SkipVerify(true),
		api.TLSMinVersion("1.2"))
	if err != nil {
		log.Panicf("can't create new target: %s", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err = t.CreateGNMIClient(ctx)
	defer t.Close()
	if err != nil {
		log.Panicf("unable create gNMI client: %s", err)
	}

	getSystem, err := api.NewGetRequest(api.Path("/system"),
		//api.Prefix("/"),
		api.DataTypeCONFIG(),
		api.EncodingJSON_IETF())

	if err != nil {
		log.Panicf("can't construct gNMI Get request: %s", err)
	}

	fmt.Println(prototext.Format(getSystem))

	respSystem, err := t.Get(ctx, getSystem)
	if err != nil {
		log.Panicf("error sending gNMI Get request: %s", err)
	}
	bSysConf := respSystem.Notification[0].Update[0].Val.GetJsonIetfVal()
	fmt.Println(string(bSysConf))

	// Populating model

	srl := new(ysrl.Device)
	srl.System = new(ysrl.SrlNokiaSystem_System)

	if err = ysrl.Unmarshal(bSysConf, srl.System); err != nil {
		log.Fatalln(err)
	}
	// ygot system lib

	fmt.Println(*srl.System.Banner.LoginBanner)
	//fmt.Println(*srl.System.GetName())

	if srl.System.GetName() == nil {
		srl.System.Name = new(ysrl.SrlNokiaSystem_System_Name)
		srl.System.Name.HostName = ygot.String("clab-2nd-srl1")
	} else {
		if *srl.System.Name.HostName == "clab-2nd-srl1" {
			srl.System.Name.HostName = ygot.String("clab-2nd-srl1-test")
		}
	}

	if err := srl.System.Validate(); err != nil {
		log.Panicf("unable to validate: %s", err)
	}
	sysMapJSON, err := ygot.ConstructIETFJSON(srl.System, nil)
	if err != nil {
		log.Panicf("unable to marshall into map:  %s", err)
	}
	// sysJSON, err := json.MarshalIndent(sysMapJSON, "", "    ")
	// if err != nil {
	// 	log.Panicf("unable to marshall into JSON:  %s", err)
	// }

	updSystem, err := api.NewSetRequest(api.Update(api.Path("/system"), api.Value(sysMapJSON, "json_ietf")))
	if err != nil {
		log.Panicf("can't construct gNMI Set request: %s", err)
	}

	fmt.Printf("GetPrefix(): %s", updSystem.GetPrefix())
	updSystemResp, err := t.Set(ctx, updSystem)
	if err != nil {
		log.Panicf("err sending set requst:  %s", err)
	}
	fmt.Println(updSystemResp.String())

	//srl.System.Name.HostName = ygot.String("ygot-pg")
	// srl.System.Name.DomainName = ygot.String("ygot.com")

	// srl.System.Dns.ServerList = append(srl.System.Dns.ServerList, "172.22.1.1")

	// if err := srl.System.Validate(); err != nil {
	// 	log.Panicf("unable to validate: %s", err)
	// }

	// Template generation
	srl = new(ysrl.Device)
	srl.System = new(ysrl.SrlNokiaSystem_System)
	srl.System.Name = new(ysrl.SrlNokiaSystem_System_Name)
	srl.System.Dns = new(ysrl.SrlNokiaSystem_System_Dns)
	srl.System.Name.HostName = ygot.String("ygot-pg")
	srl.System.Name.DomainName = ygot.String("ygot.com")

	srl.System.Dns.ServerList = append(srl.System.Dns.ServerList, "172.22.1.1")
	if err := srl.System.Validate(); err != nil {
		log.Panicf("unable to validate: %s", err)
	}

	mapJSON, err := ygot.ConstructIETFJSON(srl, nil)
	if err != nil {
		log.Panicf("unable to marshall into map:  %s", err)
	}
	json, err := json.MarshalIndent(mapJSON, "", "    ")
	fmt.Println(string(json))

}

// Localy generated
