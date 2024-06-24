// Copyright 2023 The concrete-geth Authors
//
// The concrete-geth library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The concrete library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the concrete library. If not, see <http://www.gnu.org/licenses/>.

package testtool

import (
	"bytes"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/ethereum/go-ethereum/concrete"
)

func forgeBuild(t *testing.T) error {
	cmd := exec.Command("forge", "build")
	var stdout, stderr bytes.Buffer
	cmd.Dir = "./testdata"
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	t.Log(stdout.String())
	t.Log(stderr.String())
	return nil
}

func TestRunTestContract(t *testing.T) {
	if err := forgeBuild(t); err != nil {
		t.Fatal(err)
	}
	config := TestConfig{
		Contract: filepath.Join("testdata", "src", "Test.sol:Test"),
		TestDir:  filepath.Join("testdata", "src"),
		OutDir:   filepath.Join("testdata", "out"),
	}
	concreteRegistry := concrete.NewRegistry()
	Test(t, concreteRegistry, config)
}
