package bmap

import (
	"crypto/sha1"
	"crypto/sha256"
	"fmt"
	"strings"
	"testing"
)

func TestParseIntegrity(t *testing.T) {
	templateBmap := `<?xml version="1.0" ?>
<bmap version="2.0">
    <ImageSize> 821752 </ImageSize>
    <BlockSize> 4096 </BlockSize>
    <BlocksCount> 201 </BlocksCount>
    <MappedBlocksCount> 117 </MappedBlocksCount>
    <ChecksumType> sha256 </ChecksumType>
    <BmapFileChecksum> PLACEHOLDER </BmapFileChecksum>
    <BlockMap>
        <Range chksum="9eaf19215d55d23de1be1fe4bed4a95bfe620a404352fd06e782738fff58e500"> 0-1 </Range>
    </BlockMap>
</bmap>`

	// Calculate valid checksum for the template
	zeroed := strings.Replace(templateBmap, "PLACEHOLDER", strings.Repeat("0", 64), 1)
	h := sha256.New()
	h.Write([]byte(zeroed))
	validChecksum := fmt.Sprintf("%x", h.Sum(nil))
	validBmap := strings.Replace(templateBmap, "PLACEHOLDER", validChecksum, 1)

	t.Run("ValidBmap", func(t *testing.T) {
		r := strings.NewReader(validBmap)
		b, err := Parse(r)
		if err != nil {
			t.Fatalf("Parse failed: %v", err)
		}
		if b.BmapFileChecksum != validChecksum {
			t.Errorf("unexpected checksum: %s", b.BmapFileChecksum)
		}
	})

	t.Run("CorruptedBmap", func(t *testing.T) {
		// Change one character in the XML that is NOT the checksum itself
		corruptedBmap := strings.Replace(validBmap, "<BlocksCount> 201 </BlocksCount>", "<BlocksCount> 202 </BlocksCount>", 1)
		r := strings.NewReader(corruptedBmap)
		_, err := Parse(r)
		if err == nil {
			t.Fatal("expected error for corrupted bmap, but got nil")
		}
		if !strings.Contains(err.Error(), "checksum mismatch") {
			t.Errorf("expected checksum mismatch error, got: %v", err)
		}
	})

    t.Run("ModifiedChecksum", func(t *testing.T) {
        // If we modify the checksum string in the XML, it should fail because the calculated hash (of the XML with '0's) won't match the new string
		// Change the first character of the checksum
		newChecksum := "a" + validChecksum[1:]
		corruptedBmap := strings.Replace(validBmap, validChecksum, newChecksum, 1)
		r := strings.NewReader(corruptedBmap)
		_, err := Parse(r)
		if err == nil {
			t.Fatal("expected error for modified checksum, but got nil")
		}
    })
}

func TestSHA1Integrity(t *testing.T) {
    // A hypothetical SHA1 bmap (v1.3 style)
    // We need a valid SHA1 checksum for this.
    // Let's calculate one.
    content := `<?xml version="1.0" ?>
<bmap version="1.3">
    <ImageSize> 4096 </ImageSize>
    <BlockSize> 4096 </BlockSize>
    <BlocksCount> 1 </BlocksCount>
    <MappedBlocksCount> 1 </MappedBlocksCount>
    <BmapFileSHA1> PLACEHOLDER_SHA1_HERE </BmapFileSHA1>
    <BlockMap>
        <Range sha1="some-sha1"> 0 </Range>
    </BlockMap>
</bmap>`
    
    // Replace PLACEHOLDER_SHA1_HERE with 40 '0's
    zeroed := strings.Replace(content, "PLACEHOLDER_SHA1_HERE", strings.Repeat("0", 40), 1)
    
    h := sha1.New()
    h.Write([]byte(zeroed))
    actualSHA1 := fmt.Sprintf("%x", h.Sum(nil))
    
    finalContent := strings.Replace(content, "PLACEHOLDER_SHA1_HERE", actualSHA1, 1)
    
    r := strings.NewReader(finalContent)
    b, err := Parse(r)
    if err != nil {
        t.Fatalf("Parse SHA1 failed: %v", err)
    }
    if b.BmapFileSHA1 != actualSHA1 {
        t.Errorf("unexpected SHA1: %s", b.BmapFileSHA1)
    }
}
