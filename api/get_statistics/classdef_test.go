package getstatistics

import "testing"

func TestStatisticsFromJSON(t *testing.T) {
	seq := make([]any, 8)
	for i := range seq {
		seq[i] = map[string]any{
			"group": []any{
				map[string]any{},
				map[string]any{
					"values": []any{
						map[string]any{"value": i},
						map[string]any{"value": i * 10},
					},
				},
			},
		}
	}

	got := StatisticsFromJSON(seq)

	// extract(0) -> View = [0, 0]; extract(1) -> Thread = [1, 10]; etc.
	if len(got.View) != 2 || got.View[0] != 0 || got.View[1] != 0 {
		t.Errorf("View = %v", got.View)
	}
	if len(got.Thread) != 2 || got.Thread[0] != 1 || got.Thread[1] != 10 {
		t.Errorf("Thread = %v", got.Thread)
	}
	if len(got.Recommend) != 2 || got.Recommend[0] != 7 || got.Recommend[1] != 70 {
		t.Errorf("Recommend = %v", got.Recommend)
	}
}
