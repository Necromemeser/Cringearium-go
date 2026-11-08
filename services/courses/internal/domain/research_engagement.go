package domain

type ResearchEngagementMetrics struct {
	Improved        int          `json:"improved"`
	Unchanged       int          `json:"unchanged"`
	Declined        int          `json:"declined"`
	ImprovementRate *float64     `json:"improvement_rate"`
	Progress        SummaryStats `json:"progress"`
	CalendarDays    SummaryStats `json:"calendar_days"`
	AISessions      SummaryStats `json:"ai_sessions"`
	AIAnswers       SummaryStats `json:"ai_answers"`
	PreAttempts     SummaryStats `json:"pre_attempts"`
	PostAttempts    SummaryStats `json:"post_attempts"`
	AITotalAnswers  int          `json:"ai_total_answers"`
	AIAccuracy      *float64     `json:"ai_accuracy"`
}

func summarizeEngagement(arm string, rows []ResearchParticipant) ResearchEngagementMetrics {
	e := ResearchEngagementMetrics{}
	var progress, days, sessions, answers, pres, posts []float64
	correct, paired := 0, 0
	for _, r := range rows {
		if r.Arm != arm {
			continue
		}
		pres = append(pres, float64(r.PreAttempts))
		posts = append(posts, float64(r.PostAttempts))
		if r.ProgressPercent != nil {
			progress = append(progress, *r.ProgressPercent)
		}
		if r.AISessions != nil {
			sessions = append(sessions, float64(*r.AISessions))
		}
		if r.AIAnswers != nil {
			answers = append(answers, float64(*r.AIAnswers))
			e.AITotalAnswers += *r.AIAnswers
		}
		if r.AICorrect != nil {
			correct += *r.AICorrect
		}
		if r.Eligible && r.Gain != nil {
			paired++
			if *r.Gain > 0 {
				e.Improved++
			} else if *r.Gain < 0 {
				e.Declined++
			} else {
				e.Unchanged++
			}
			if r.CalendarDays != nil {
				days = append(days, *r.CalendarDays)
			}
		}
	}
	if paired > 0 {
		e.ImprovementRate = fptr(100 * float64(e.Improved) / float64(paired))
	}
	if e.AITotalAnswers > 0 {
		e.AIAccuracy = fptr(100 * float64(correct) / float64(e.AITotalAnswers))
	}
	e.Progress = Summarize(progress)
	e.CalendarDays = Summarize(days)
	e.AISessions = Summarize(sessions)
	e.AIAnswers = Summarize(answers)
	e.PreAttempts = Summarize(pres)
	e.PostAttempts = Summarize(posts)
	return e
}
