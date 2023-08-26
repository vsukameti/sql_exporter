package config

import (
	"reflect"
	"testing"
)

func TestResolveCollectorRefs(t *testing.T) {
	colls := map[string]*CollectorConfig{
		"a":  {Name: "a"},
		"b":  {Name: "b"},
		"c":  {Name: "b"},
		"aa": {Name: "aa"},
	}

	t.Run("NoGlobbing", func(t *testing.T) {
		crefs := []string{
			"a",
			"b",
		}
		cs, err := resolveCollectorRefs(crefs, colls, "target")
		if err != nil {
			t.Fatalf("expected no error but got: %v", err)
		}
		if len(cs) != 2 {
			t.Fatalf("expected len(cs)=2 but got len(cs)=%d", len(cs))
		}
		expected := []*CollectorConfig{
			colls["a"],
			colls["b"],
		}
		if !reflect.DeepEqual(cs, expected) {
			t.Fatalf("expected cs=%v but got cs=%v", expected, cs)
		}
	})

	t.Run("Globbing", func(t *testing.T) {
		crefs := []string{
			"a*",
			"b",
		}
		cs, err := resolveCollectorRefs(crefs, colls, "target")
		if err != nil {
			t.Fatalf("expected no error but got: %v", err)
		}
		if len(cs) != 3 {
			t.Fatalf("expected len(cs)=3 but got len(cs)=%d", len(cs))
		}
		expected1 := []*CollectorConfig{
			colls["a"],
			colls["aa"],
			colls["b"],
		}
		expected2 := []*CollectorConfig{ // filepath.Match() is non-deterministic
			colls["aa"],
			colls["a"],
			colls["b"],
		}
		if !reflect.DeepEqual(cs, expected1) && !reflect.DeepEqual(cs, expected2) {
			t.Fatalf("expected cs=%v or cs=%v but got cs=%v", expected1, expected2, cs)
		}
	})

	t.Run("NoCollectorRefs", func(t *testing.T) {
		crefs := []string{}
		cs, err := resolveCollectorRefs(crefs, colls, "target")
		if err != nil {
			t.Fatalf("expected no error but got: %v", err)
		}
		if len(cs) != 0 {
			t.Fatalf("expected len(cs)=0 but got len(cs)=%d", len(cs))
		}
	})

	t.Run("UnknownCollector", func(t *testing.T) {
		crefs := []string{
			"a",
			"x",
		}
		_, err := resolveCollectorRefs(crefs, colls, "target")
		if err == nil {
			t.Fatalf("expected error but got none")
		}
		// TODO: Code should use error types and check with 'errors.Is(err1, err2)'.
		expected := "unknown collector \"x\" referenced in target"
		if err.Error() != expected {
			t.Fatalf("expected err=%q but got err=%q", expected, err.Error())
		}
	})
}

// // write a test for readDSNFromAwsSecretManager function
// func TestReadDSNFromAwsSecretManager(t *testing.T) {
// 	t.Run("NoDSN", func(t *testing.T) {
// 		secret := &secretsmanager.GetSecretValueOutput{}
// 		err := readDSNFromAwsSecretManager(secret)
// 		var expected Secret = "No DSN found in secret"
// 		if err == nil {
// 			t.Fatalf("expected error but got none")
// 		}
// 		if err != expected {
// 			t.Fatalf("expected err=%q but got err=%q", expected, err)
// 		}
// 	})
// 	t.Run("DSN", func(t *testing.T) {
// 		secret := &secretsmanager.GetSecretValueOutput{
// 			SecretString: aws.String("dsn"),
// 		}
// 		dsn, err := readDSNFromAwsSecretManager(secret)
// 		if err != nil {
// 			t.Fatalf("expected no error but got: %v", err)
// 		}
// 		if dsn != "dsn" {
// 			t.Fatalf("expected dsn=%q but got dsn=%q", "dsn", dsn)
// 		}
// 	})
// }
//

// write a test for LoadCollectorFiles function
func TestLoadCollectorFiles(t *testing.T) {
	t.Run("NoFiles", func(t *testing.T) {
		config := Config{
			Globals:        &GlobalConfig{},
			CollectorFiles: []string{},
			Target:         &TargetConfig{},
			Jobs:           []*JobConfig{},
			Collectors:     []*CollectorConfig{},
			configFile:     "",
			XXX:            map[string]any{},
		}
		err := config.loadCollectorFiles()
		if err != nil {
			t.Fatalf("expected no error but got: %v", err)
		}
		//		if len(collectorConfigs) != 0 {
		//			t.Fatalf("expected len(collectorConfigs)=0 but got len(collectorConfigs)=%d", len(collectorConfigs))
		//		}
	})
	//	t.Run("Files", func(t *testing.T) {
	//		//		collectorFiles := []string{"testdata/collector.yaml"}
	//		config := Config{
	//			Globals:        &GlobalConfig{},
	//			CollectorFiles: []string{},
	//			Target:         &TargetConfig{},
	//			Jobs:           []*JobConfig{},
	//			Collectors:     []*CollectorConfig{},
	//			configFile:     "",
	//			XXX:            map[string]any{},
	//		}
	//		err := config.LoadCollectorFiles()
	//		if err != nil {
	//			t.Fatalf("expected no error but got: %v", err)
	//		}
	//		//if len(collectorConfigs) != 1 {
	//		//	t.Fatalf("expected len(collectorConfigs)=1 but got len(collectorConfigs)=%d", len(collectorConfigs))
	//		//}
	//		expected := []*CollectorConfig{
	//			{
	//				Name: "test",
	//				Queries: []Query{
	//					{
	//						Name: "test",
	//						SQL:  "SELECT 1",
	//					},
	//				},
	//			},
	//		}
	//		if !reflect.DeepEqual(collectorConfigs, expected) {
	//			t.Fatalf("expected collectorConfigs=%v but got collectorConfigs=%v", expected, collectorConfigs)
	//		}
	//	})
}
