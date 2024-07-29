package drivers

import (
	"context"
	"math"
	"time"

	"github.com/valkey-io/valkey-go"
)

type valkeyDriver struct {
	client valkey.Client
	prefix string
}

func NewValkeyDriver(client valkey.Client) Driver {
	return &valkeyDriver{client: client, prefix: "stash:"}
}

func (d *valkeyDriver) SetPrefix(prefix string) Driver {
	d.prefix = prefix
	return d
}

func (d *valkeyDriver) Init() error { return nil }

func (d *valkeyDriver) Forget(ctx context.Context, key string) (bool, error) {
	cmd := d.client.B().Del().Key(d.prefixedKey(key)).Build()
	resp := d.client.Do(ctx, cmd)

	return resp.AsBool()
}

func (d *valkeyDriver) Add(ctx context.Context, raw RawValue) error {
	lua := valkey.NewLuaScript("return valkey.call('exists', KEYS[1]) > 1 and valkey.call('set', KEYS[1], ARGV[1], 'EX', ARGV[2])")
	return lua.Exec(ctx, d.client, []string{d.prefixedKey(raw.Key)}, []string{raw.AsString(), raw.Expires.Sub(time.Now()).String()}).Error()
}

func (d *valkeyDriver) Flush(ctx context.Context) error {
	cmd := d.client.B().Flushdb().Build()
	return d.client.Do(ctx, cmd).Error()
}

func (d *valkeyDriver) Forever(ctx context.Context, raw RawValue) error {
	cmd := d.client.B().Set().Key(d.prefixedKey(raw.Key)).Value(raw.AsString()).Build()
	return d.client.Do(ctx, cmd).Error()
}

func (d *valkeyDriver) Get(ctx context.Context, key string) (*RawValue, error) {
	cmd := d.client.B().Get().Key(d.prefixedKey(key)).Build()
	resp := d.client.Do(ctx, cmd)
	if err := resp.Error(); err != nil {
		if valkey.IsValkeyNil(err) {
			return nil, nil
		}

		return nil, err
	}

	b, err := resp.AsBytes()
	if err != nil {
		return nil, err
	}

	return &RawValue{Key: key, Value: b, Expires: time.Time{}}, nil
}

func (d *valkeyDriver) Put(ctx context.Context, raw RawValue) error {
	seconds := int64(math.Max(1, raw.Expires.Sub(time.Now()).Seconds()))
	cmd := d.client.B().Setex().Key(d.prefixedKey(raw.Key)).Seconds(seconds).Value(raw.AsString()).Build()

	return d.client.Do(ctx, cmd).Error()
}

func (d *valkeyDriver) prefixedKey(key string) string {
	return d.prefix + key
}
