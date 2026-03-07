package e2e

import (
	"fmt"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/stretchr/testify/require"

	"go.etcd.io/etcd/client/pkg/v3/fileutil"
	"go.etcd.io/etcd/tests/v3/framework/config"
	"go.etcd.io/etcd/tests/v3/framework/e2e"
)

func TestRollingReplacementUpgrade(t *testing.T) {
	// Check if last release binary exists
	lastReleaseBinary := e2e.BinPath.EtcdLastRelease
	// fi, err := os.Stat(lastReleaseBinary)
	// fmt.Printf("fi, %v, err, %v\n", fi, err)
	//
	// es, err := os.ReadDir("./")
	// fmt.Printf("es, %v, err, %v\n", es, err)
	//
	// os.Gwd()
	// es, err := os.Pwd("./bin")
	// fmt.Printf("es, %v, err, %v\n", es, err)

	if !fileutil.Exist(lastReleaseBinary) {
		t.Skipf("%q does not exist", lastReleaseBinary)
	}
	currentEtcdBinary := e2e.BinPath.Etcd
	currentVersion, err := e2e.GetVersionFromBinary(currentEtcdBinary)
	require.NoError(t, err)
	currentVersion.PreRelease = ""
	lastVersion, err := e2e.GetVersionFromBinary(lastReleaseBinary)
	require.NoError(t, err)

	lastClusterVersion := semver.New(lastVersion.String())
	lastClusterVersion.Patch = 0
	e2e.BeforeTest(t)
	// Create cluster with 3 members
	epc, err := e2e.NewEtcdProcessCluster(t.Context(), t,
		e2e.WithClusterSize(3),
		e2e.WithSnapshotCount(10),
		e2e.WithKeepDataDir(true),
	)
	require.NoError(t, err)
	t.Cleanup(func() {
		epc.Close()
	})

	// Optional: write some data before upgrade
	cc := epc.Etcdctl()
	ctx := t.Context()
	for i := 0; i < 10; i++ {
		cc.Put(ctx, fmt.Sprintf("key-%d", i), "value", config.PutOptions{})
	}

	// Run rolling replacement upgrade
	err = e2e.RollingReplacementUpgrade(t, nil, epc, lastClusterVersion, currentVersion)
	require.NoError(t, err)

	// Verify cluster still works
	resp, err := cc.Get(ctx, "key-", config.GetOptions{Prefix: true})
	require.NoError(t, err)
	require.Equal(t, 10, len(resp.Kvs))
}
