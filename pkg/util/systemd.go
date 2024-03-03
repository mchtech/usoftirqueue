package util

import (
	"context"

	"github.com/coreos/go-systemd/v22/dbus"
	"github.com/pkg/errors"
)

type SystemCtl struct {
	conn *dbus.Conn
	ctx  context.Context
}

func NewWithContext(ctx context.Context) (*SystemCtl, error) {
	conn, err := dbus.NewWithContext(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to system bus")
	}
	return &SystemCtl{conn: conn, ctx: ctx}, nil
}

func (s *SystemCtl) Close() {
	s.conn.Close()
}

func (s *SystemCtl) StopService(servicename string) (err error) {
	if _, err := s.conn.StopUnitContext(s.ctx, servicename, "replace", nil); err != nil {
		return errors.Wrap(err, "failed to stop unit")
	}
	return
}

func (s *SystemCtl) StartService(servicename string) (err error) {
	if _, err := s.conn.StartUnitContext(s.ctx, servicename, "replace", nil); err != nil {
		return errors.Wrap(err, "failed to start unit")
	}
	return
}
