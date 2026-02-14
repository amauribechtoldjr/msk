package encryption

import "testing"

func TestRandomBytes(t *testing.T) {
	t.Run("should return the bytes size correctly", func(t *testing.T) {
		expectedSizes := []int{0, 12, 9, 16, 99, 9840}

		for _, expectedSize := range expectedSizes {
			bytesArray, err := randomBytes(expectedSize)
			if err != nil {
				t.Fatal("failed to create the bytes array")
			}

			if len(bytesArray) != expectedSize {
				t.Fatalf("expected array size of %v, and got %v", expectedSize, len(bytesArray))
			}
		}
	})
}
