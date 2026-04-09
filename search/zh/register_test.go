package zh

import "testing"

func TestDictFilesFromConfig_JSONNull(t *testing.T) {
	// Bleve 持久化 mapping 再打开时，dict_files 常为 JSON null
	got, err := dictFilesFromConfig(map[string]interface{}{"dict_files": nil})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("want empty slice, got %#v", got)
	}
}
