package upnp

import (
	"mapserver/app"
	"strconv"
	"fmt"
	"net"
	"time"
	"log"
	"net/url"
	"net/http"
	"strings"
	"math/rand"
)

const(
	rootDescPath       = "/upnp/"
	rootDeviceType     = "urn:evidenceb:device:Mapserver:1"
	ssdpInterfaceFlags = net.FlagUp | net.FlagMulticast
)

type UpnpServer struct {
	HTTPConn       net.Listener
	Interfaces     []net.Interface
	Port           int
}

var UUID string;

// Generate one UUID for this mapserver instance
func init() {
	b:= make([]byte, 16)
	rand.Seed(time.Now().UTC().UnixNano())
	rand.Read(b)
	b[8] = b[8] & 0x0f | 0x40 // Version 4 : Random UUID
	b[10] = b[10] & 0x3f | 0x80 // Variant 1 : ?

	UUID = fmt.Sprintf(
		"%02x%02x%02x%02x-%02x%02x-%02x%02x-%02x%02x-%02x%02x%02x%02x%02x%02x",
		b[0], b[1], b[2], b[3], b[4], b[5], b[6], b[7],
		b[8], b[9], b[10], b[11], b[12], b[13], b[14], b[15])
}

// https://github.com/anacrolix/dms/blob/master/dlna/dms/dms.go

func (this *UpnpServer)getLocation(ip net.IP) string {
	url := url.URL{
		Scheme: "http",
		Host: (&net.TCPAddr{
			IP:   ip,
			Port: this.Port,
		}).String(),
		Path: rootDescPath + ip.String(),
	}
	return url.String()
}

func (this *UpnpServer)serveSSDP(intf net.Interface) {
	s := Server{
		Interface: intf,
		Devices:   []string{ rootDeviceType },
		Services:  []string{}, // No service in Kidscode
		Location: func(ip net.IP) string {
			return this.getLocation(ip)
		},
		UUID:           UUID,
		NotifyInterval: 30*time.Second,
	}
	if err := s.Init(); err != nil {
		if intf.Flags&ssdpInterfaceFlags != ssdpInterfaceFlags {
			// Didn't expect it to work anyway.
			return
		}
		if strings.Contains(err.Error(), "listen") {
			// OSX has a lot of dud interfaces. Failure to create a socket on
			// the interface are what we're expecting if the interface is no
			// good.
			return
		}
		log.Printf("error creating ssdp server on %s: %s", intf.Name, err)
		return
	}
	defer s.Close()
	log.Println("UPNP: Announce on", intf.Name)
	stopped := make(chan struct{})
	go func() {
		defer close(stopped)
		if err := s.Serve(); err != nil {
			log.Printf("%q: %q\n", intf.Name, err)
		}
	}()
	select {
//	case <-me.closed:
		// Returning will close the server.
	case <-stopped:
	}
}

type UpnpHandler struct {
	Ctx *app.App
}

func (this *UpnpHandler)getXML(ip string) []byte {
	return []byte("<?xml version=\"1.0\"?>" +
		"<root xmlns=\"urn:schemas-upnp-org:device-1-0\" xmlns:kc=\"urn:schemas-kidscode-org:server-1-0\">" +
		"<specVersion><major>1</major><minor>0</minor></specVersion>" +
		"<device>" +
			"<deviceType>" + rootDeviceType + "</deviceType>" +
			"<friendlyName>Kidscode</friendlyName>" +
			"<manufacturer>EvidenceB</manufacturer>" +
			"<modelName>Kidscode</modelName>" +
			"<UDN>uuid:" + UUID + "</UDN>" +
		"</device>" +
		// Kidscode specific part
		"<kc:mapserver>" +
			"<kc:url>http://" + ip + ":" + strconv.Itoa(this.Ctx.Config.Port) + "</kc:url>" +
			"<kc:name>" + "" + "</kc:name>" +
			"<kc:version>" + "" + "</kc:version>" +
		"</kc:mapserver>" +
	"</root>")
}

func (this *UpnpHandler) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	str := strings.TrimPrefix(req.URL.Path, "/upnp/")
	var xml []byte = this.getXML(str)
	resp.Header().Set("Ext", "")
	resp.Header().Set("content-type", `text/xml; charset="utf-8"`)
	resp.Header().Set("content-length", fmt.Sprint(len(xml)))
	resp.Write(xml)
}

func (this *UpnpServer) Start(ctx *app.App) {
	this.HTTPConn, _ = net.Listen("tcp", "")
	this.Port = ctx.Config.Port
	itfs, _ := net.Interfaces()

	// Announce on all non loopback interfaces
	for _, itf := range itfs {
		if (itf.Flags&net.FlagLoopback == net.FlagLoopback) {
			continue;
		}
		go func() {
			this.serveSSDP(itf)
		} ();
	}
}

func Announce(ctx *app.App) {
	server := new(UpnpServer)
	server.Start(ctx)
}
