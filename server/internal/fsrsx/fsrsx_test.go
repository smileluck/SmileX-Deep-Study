package fsrsx

import (
	"errors"
	"testing"
	"time"
)

var testNow = time.Date(2026, 9, 16, 8, 0, 0, 0, time.UTC)

func TestFromMapToMapRoundTrip(t *testing.T) {
	st := New(testNow)
	m := st.ToMap()
	back, err := FromMap(m)
	if err != nil {
		t.Fatalf("FromMap: %v", err)
	}
	if !back.Due.Equal(testNow) {
		t.Errorf("Due = %v, want %v", back.Due, testNow)
	}
	if back.Reps != 0 || back.State != 0 || back.LastReview != nil {
		t.Errorf("零状态走样: %+v", back)
	}
	// 评过一次后带 last_review 再往返
	next, err := Grade(st, 3, testNow)
	if err != nil {
		t.Fatalf("Grade: %v", err)
	}
	back2, err := FromMap(next.ToMap())
	if err != nil {
		t.Fatalf("FromMap(graded): %v", err)
	}
	if back2.Reps != 1 || back2.LastReview == nil || !back2.LastReview.Equal(testNow) {
		t.Errorf("评分后往返走样: %+v", back2)
	}
	if back2.Stability != next.Stability || back2.Difficulty != next.Difficulty {
		t.Errorf("stability/difficulty 不一致: %+v vs %+v", back2, next)
	}
}

func TestGradeNewCard(t *testing.T) {
	// go-fsrs v1.2.1：新卡 Again/Hard/Good → Learning(1)，Easy → Review(2)
	wantState := map[int]int{1: 1, 2: 1, 3: 1, 4: 2}
	for rating := 1; rating <= 4; rating++ {
		next, err := Grade(New(testNow), rating, testNow)
		if err != nil {
			t.Fatalf("Grade(%d): %v", rating, err)
		}
		if next.Reps != 1 {
			t.Errorf("rating=%d Reps = %d, want 1", rating, next.Reps)
		}
		if next.LastReview == nil || !next.LastReview.Equal(testNow) {
			t.Errorf("rating=%d LastReview = %v", rating, next.LastReview)
		}
		if next.State != wantState[rating] {
			t.Errorf("rating=%d State = %d, want %d", rating, next.State, wantState[rating])
		}
		if next.Due.Before(testNow) {
			t.Errorf("rating=%d Due 早于评分时间: %v", rating, next.Due)
		}
	}
	// Easy 的到期应远于 Again
	again, _ := Grade(New(testNow), 1, testNow)
	easy, _ := Grade(New(testNow), 4, testNow)
	if !easy.Due.After(again.Due) {
		t.Errorf("Easy due %v 应晚于 Again due %v", easy.Due, again.Due)
	}
}

func TestGradeStateTransitions(t *testing.T) {
	// New →(Good)→ Learning →(Good)→ Review →(Again)→ Relearning
	st, err := Grade(New(testNow), 3, testNow)
	if err != nil || st.State != 1 {
		t.Fatalf("Good on new: state=%d err=%v", st.State, err)
	}
	later := testNow.Add(time.Hour)
	st, err = Grade(st, 3, later)
	if err != nil || st.State != 2 {
		t.Fatalf("Good on learning: state=%d err=%v", st.State, err)
	}
	st, err = Grade(st, 1, later.Add(time.Hour))
	if err != nil || st.State != 3 {
		t.Fatalf("Again on review: state=%d err=%v", st.State, err)
	}
	if st.Reps != 3 {
		t.Errorf("Reps = %d, want 3", st.Reps)
	}
}

func TestGradeInvalidRating(t *testing.T) {
	for _, r := range []int{0, 5, -1} {
		if _, err := Grade(New(testNow), r, testNow); !errors.Is(err, ErrInvalidRating) {
			t.Errorf("rating=%d err = %v, want ErrInvalidRating", r, err)
		}
	}
}

func TestStateName(t *testing.T) {
	want := map[int]string{0: "new", 1: "learning", 2: "review", 3: "relearning", 9: "unknown"}
	for s, name := range want {
		if got := StateName(s); got != name {
			t.Errorf("StateName(%d) = %s, want %s", s, got, name)
		}
	}
}
