package pantavisor

import (
	"encoding/json"
	"testing"
)

func TestUnmarshalReleases_NewSchema(t *testing.T) {
	data := []byte(`{
		"stable": {
			"029": {
				"docs": {
					"name": "pantavisor-raspberrypi4-64-20260601.docs.tar.zst",
					"hash": "e3b0c44298fc",
					"url": "https://example.com/029/docs.tar.zst"
				},
				"devices": [
					{
						"name": "docker-x86_64-scarthgap",
						"full_image": { "url": "https://example.com/full.tar.gz", "sha256": "abc123" },
						"pvrexports": { "url": "https://example.com/pvexports.tar.gz", "sha256": "def456" },
						"bsp": { "url": "https://example.com/bsp.tgz", "sha256": "789" },
						"sdk": { "url": "https://example.com/sdk.sh", "sha256": "555" }
					},
					{ "name": "raspberrypi-armv8-scarthgap", "full_image": { "url": "u", "sha256": "s" } },
					{ "timestamp": "2026-06-03T10:00+00:00" }
				]
			}
		},
		"release-candidate": {
			"029-rc1": {
				"docs": { "name": "d", "hash": "h", "url": "u" },
				"devices": [ { "name": "docker-x86_64-scarthgap", "full_image": { "url": "u", "sha256": "s" } } ]
			}
		}
	}`)

	var releases Releases
	if err := json.Unmarshal(data, &releases); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	rw := releases["stable"]["029"]
	if rw.Docs == nil {
		t.Fatal("expected docs to be parsed")
	}
	if rw.Docs.Hash != "e3b0c44298fc" {
		t.Errorf("docs hash = %q", rw.Docs.Hash)
	}
	if rw.Timestamp != "2026-06-03T10:00+00:00" {
		t.Errorf("timestamp = %q, want it extracted from devices marker", rw.Timestamp)
	}
	if len(rw.Devices) != 2 {
		t.Fatalf("expected 2 devices (timestamp marker excluded), got %d", len(rw.Devices))
	}
	d0 := rw.Devices[0]
	if d0.Name != "docker-x86_64-scarthgap" {
		t.Errorf("device[0].Name = %q", d0.Name)
	}
	if d0.FullImage.URL != "https://example.com/full.tar.gz" || d0.FullImage.SHA256 != "abc123" {
		t.Errorf("device[0].FullImage = %+v", d0.FullImage)
	}
	if d0.SDK.URL != "https://example.com/sdk.sh" {
		t.Errorf("device[0].SDK = %+v", d0.SDK)
	}
}

func TestUnmarshalReleases_LegacySchema(t *testing.T) {
	// Current live schema: release-date sibling key, no docs.
	data := []byte(`{
		"stable": {
			"024": {
				"release-date": "2026-02-13T16:19+00:00",
				"devices": [
					{ "name": "docker-x86_64-scarthgap", "full_image": { "url": "u", "sha256": "s" } }
				]
			}
		}
	}`)

	var releases Releases
	if err := json.Unmarshal(data, &releases); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	rw := releases["stable"]["024"]
	if rw.Docs != nil {
		t.Errorf("expected no docs, got %+v", rw.Docs)
	}
	if rw.Timestamp != "2026-02-13T16:19+00:00" {
		t.Errorf("timestamp = %q, want release-date fallback", rw.Timestamp)
	}
	if len(rw.Devices) != 1 || rw.Devices[0].Name != "docker-x86_64-scarthgap" {
		t.Errorf("devices = %+v", rw.Devices)
	}
}

func TestUnmarshalReleases_DocsOnly(t *testing.T) {
	// A version may carry only docs and no devices (seen live as unknown/028-rc12).
	data := []byte(`{
		"unknown": {
			"028-rc12": { "docs": { "name": "d", "hash": "h", "url": "u" } }
		}
	}`)

	var releases Releases
	if err := json.Unmarshal(data, &releases); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	rw := releases["unknown"]["028-rc12"]
	if rw.Docs == nil || rw.Docs.Name != "d" {
		t.Errorf("docs = %+v", rw.Docs)
	}
	if len(rw.Devices) != 0 {
		t.Errorf("expected 0 devices, got %d", len(rw.Devices))
	}
}

func TestUnmarshalReleases_BareListSchema(t *testing.T) {
	data := []byte(`{
		"stable": {
			"023": [
				{ "name": "docker-x86_64-scarthgap", "full_image": { "url": "u", "sha256": "s" } }
			]
		}
	}`)

	var releases Releases
	if err := json.Unmarshal(data, &releases); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	rw := releases["stable"]["023"]
	if len(rw.Devices) != 1 || rw.Devices[0].Name != "docker-x86_64-scarthgap" {
		t.Errorf("devices = %+v", rw.Devices)
	}
}
