package alloc

import (
	"fmt"
	"net"

	"github.com/flynn/flynn/flannel/backend"
	"github.com/flynn/flynn/flannel/pkg/ip"
	"github.com/flynn/flynn/flannel/pkg/task"
	"github.com/flynn/flynn/flannel/subnet"
)

type AllocBackend struct {
	sm   *subnet.SubnetManager
	stop chan bool
}

func New(sm *subnet.SubnetManager) backend.Backend {
	return &AllocBackend{
		sm:   sm,
		stop: make(chan bool),
	}
}

func (m *AllocBackend) Init(extIface *net.Interface, extIP net.IP, httpPort string, ipMasq bool) (*backend.SubnetDef, error) {
	attrs := subnet.LeaseAttrs{
		PublicIP: ip.FromIP(extIP),
		HTTPPort: httpPort,
	}

	sn, err := m.sm.AcquireLease(&attrs, m.stop)
	if err != nil {
		if err == task.ErrCanceled {
			return nil, err
		} else {
			return nil, fmt.Errorf("failed to acquire lease: %v", err)
		}
	}

	return &backend.SubnetDef{
		Net: sn,
		MTU: extIface.MTU,
	}, nil
}

func (m *AllocBackend) Run() {
	m.sm.LeaseRenewer(m.stop)
}

func (m *AllocBackend) Stop() {
	close(m.stop)
}

func (m *AllocBackend) Name() string {
	return "allocation"
}
