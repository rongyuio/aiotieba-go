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

	// extract(0) -> View = [0, 0]; extract(1) -> Thread = [10, 1]; 服务端按时间新到旧下发，结果反转为旧到新。
	if len(got.View) != 2 || got.View[0] != 0 || got.View[1] != 0 {
		t.Errorf("View = %v", got.View)
	}
	if len(got.Thread) != 2 || got.Thread[0] != 10 || got.Thread[1] != 1 {
		t.Errorf("Thread = %v", got.Thread)
	}
	if len(got.Recommend) != 2 || got.Recommend[0] != 70 || got.Recommend[1] != 7 {
		t.Errorf("Recommend = %v", got.Recommend)
	}
}
