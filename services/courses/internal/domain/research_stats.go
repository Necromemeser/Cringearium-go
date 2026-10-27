package domain

import (
	"math"
	"sort"
)

type SummaryStats struct {
	N      int      `json:"n"`
	Mean   *float64 `json:"mean"`
	Median *float64 `json:"median"`
	SD     *float64 `json:"sd"`
	CILow  *float64 `json:"ci_low"`
	CIHigh *float64 `json:"ci_high"`
}
type ResearchGroupMetrics struct {
	Arm            string       `json:"arm"`
	Assigned       int          `json:"assigned"`
	WithPre        int          `json:"with_pre"`
	WithPost       int          `json:"with_post"`
	Paired         int          `json:"paired"`
	FinishedCourse int          `json:"finished_course"`
	PostCompletion *float64     `json:"post_completion"`
	Pre            SummaryStats `json:"pre"`
	Post           SummaryStats `json:"post"`
	Gain           SummaryStats `json:"gain"`
}
type ResearchComparison struct {
	GainDifference       *float64 `json:"gain_difference"`
	CILow                *float64 `json:"ci_low"`
	CIHigh               *float64 `json:"ci_high"`
	HedgesG              *float64 `json:"hedges_g"`
	CompletionDifference *float64 `json:"completion_difference"`
}
type RegressionResult struct {
	Available        bool     `json:"available"`
	Reason           string   `json:"reason"`
	N                int      `json:"n"`
	DF               int      `json:"df"`
	GroupCoefficient *float64 `json:"group_coefficient"`
	PreCoefficient   *float64 `json:"pre_coefficient"`
	Intercept        *float64 `json:"intercept"`
	SE               *float64 `json:"se"`
	CILow            *float64 `json:"ci_low"`
	CIHigh           *float64 `json:"ci_high"`
	R2               *float64 `json:"r2"`
}

func fptr(v float64) *float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return nil
	}
	return &v
}
func T95(df float64) float64 {
	table := []float64{0, 12.706, 4.303, 3.182, 2.776, 2.571, 2.447, 2.365, 2.306, 2.262, 2.228, 2.201, 2.179, 2.160, 2.145, 2.131, 2.120, 2.110, 2.101, 2.093, 2.086, 2.080, 2.074, 2.069, 2.064, 2.060, 2.056, 2.052, 2.048, 2.045, 2.042}
	if df < 1 {
		return table[1]
	}
	if df <= 30 {
		return table[int(math.Floor(df))]
	}
	z := 1.959963984540054
	return z + (z*z*z+z)/(4*df) + (5*math.Pow(z, 5)+16*z*z*z+3*z)/(96*df*df)
}
func Summarize(values []float64) SummaryStats {
	result := SummaryStats{N: len(values)}
	if len(values) == 0 {
		return result
	}
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean := sum / float64(len(values))
	result.Mean = fptr(mean)
	sorted := append([]float64{}, values...)
	sort.Float64s(sorted)
	m := len(sorted) / 2
	median := sorted[m]
	if len(sorted)%2 == 0 {
		median = (sorted[m-1] + median) / 2
	}
	result.Median = fptr(median)
	if len(values) < 2 {
		return result
	}
	squares := 0.0
	for _, v := range values {
		squares += (v - mean) * (v - mean)
	}
	sd := math.Sqrt(squares / float64(len(values)-1))
	result.SD = fptr(sd)
	delta := T95(float64(len(values)-1)) * sd / math.Sqrt(float64(len(values)))
	result.CILow = fptr(mean - delta)
	result.CIHigh = fptr(mean + delta)
	return result
}
func groupMetrics(arm string, rows []ResearchParticipant) ResearchGroupMetrics {
	result := ResearchGroupMetrics{Arm: arm}
	var pre, post, gain []float64
	for _, r := range rows {
		if r.Arm != arm {
			continue
		}
		result.Assigned++
		if r.PreScore != nil {
			result.WithPre++
		}
		if r.PostScore != nil {
			result.WithPost++
		}
		if r.TotalPages > 0 && r.CompletedPages >= r.TotalPages {
			result.FinishedCourse++
		}
		if r.Eligible {
			result.Paired++
			pre = append(pre, *r.PreScore)
			post = append(post, *r.PostScore)
			gain = append(gain, *r.Gain)
		}
	}
	if result.Assigned > 0 {
		result.PostCompletion = fptr(100 * float64(result.WithPost) / float64(result.Assigned))
	}
	result.Pre = Summarize(pre)
	result.Post = Summarize(post)
	result.Gain = Summarize(gain)
	return result
}
func compareGroups(groups []ResearchGroupMetrics) ResearchComparison {
	a, b := groups[1], groups[0]
	result := ResearchComparison{}
	if a.PostCompletion != nil && b.PostCompletion != nil {
		result.CompletionDifference = fptr(*a.PostCompletion - *b.PostCompletion)
	}
	if a.Gain.Mean == nil || b.Gain.Mean == nil {
		return result
	}
	diff := *a.Gain.Mean - *b.Gain.Mean
	result.GainDifference = fptr(diff)
	if a.Gain.SD == nil || b.Gain.SD == nil {
		return result
	}
	na, nb := float64(a.Paired), float64(b.Paired)
	va, vb := *a.Gain.SD**a.Gain.SD, *b.Gain.SD**b.Gain.SD
	se2 := va/na + vb/nb
	if se2 > 0 {
		df := se2 * se2 / (math.Pow(va/na, 2)/(na-1) + math.Pow(vb/nb, 2)/(nb-1))
		delta := T95(df) * math.Sqrt(se2)
		result.CILow = fptr(diff - delta)
		result.CIHigh = fptr(diff + delta)
	} else {
		result.CILow = fptr(diff)
		result.CIHigh = fptr(diff)
	}
	df := na + nb - 2
	pooled := math.Sqrt(((na-1)*va + (nb-1)*vb) / df)
	if pooled > 0 {
		result.HedgesG = fptr((1 - 3/(4*df-1)) * diff / pooled)
	}
	return result
}

// OLS post ~ pre + assigned arm with HC3 heteroskedasticity-robust covariance.
// Complete cases only; adjustment does not make observational assignment causal.
func FitANCOVA(rows []ResearchParticipant) RegressionResult {
	result := RegressionResult{Reason: "insufficient_data"}
	var x [][3]float64
	var y []float64
	groups := [2]int{}
	for _, r := range rows {
		if !r.Eligible {
			continue
		}
		g := 0
		if r.Arm == "adaptive" {
			g = 1
		}
		groups[g]++
		x = append(x, [3]float64{1, *r.PreScore, float64(g)})
		y = append(y, *r.PostScore)
	}
	n := len(x)
	result.N = n
	result.DF = n - 3
	if n < 10 || groups[0] < 3 || groups[1] < 3 {
		return result
	}
	var xtx [3][3]float64
	var xty [3]float64
	for i, row := range x {
		for j := 0; j < 3; j++ {
			xty[j] += row[j] * y[i]
			for k := 0; k < 3; k++ {
				xtx[j][k] += row[j] * row[k]
			}
		}
	}
	inverse, ok := invert3(xtx)
	if !ok {
		result.Reason = "singular_design"
		return result
	}
	var beta [3]float64
	for j := 0; j < 3; j++ {
		for k := 0; k < 3; k++ {
			beta[j] += inverse[j][k] * xty[k]
		}
	}
	var meat [3][3]float64
	residualSum, meanY := 0.0, 0.0
	for _, v := range y {
		meanY += v
	}
	meanY /= float64(n)
	totalSum := 0.0
	for i, row := range x {
		prediction, h := 0.0, 0.0
		for j := 0; j < 3; j++ {
			prediction += beta[j] * row[j]
			for k := 0; k < 3; k++ {
				h += row[j] * inverse[j][k] * row[k]
			}
		}
		if 1-h < 1e-8 {
			result.Reason = "high_leverage"
			return result
		}
		residual := y[i] - prediction
		residualSum += residual * residual
		totalSum += (y[i] - meanY) * (y[i] - meanY)
		weight := residual * residual / ((1 - h) * (1 - h))
		for j := 0; j < 3; j++ {
			for k := 0; k < 3; k++ {
				meat[j][k] += weight * row[j] * row[k]
			}
		}
	}
	variance := 0.0
	for j := 0; j < 3; j++ {
		for k := 0; k < 3; k++ {
			variance += inverse[2][j] * meat[j][k] * inverse[k][2]
		}
	}
	if variance < 0 && variance > -1e-8 {
		variance = 0
	}
	if variance < 0 {
		result.Reason = "numerical_error"
		return result
	}
	se := math.Sqrt(variance)
	delta := T95(float64(result.DF)) * se
	result.Available = true
	result.Reason = ""
	result.Intercept = fptr(beta[0])
	result.PreCoefficient = fptr(beta[1])
	result.GroupCoefficient = fptr(beta[2])
	result.SE = fptr(se)
	result.CILow = fptr(beta[2] - delta)
	result.CIHigh = fptr(beta[2] + delta)
	if totalSum > 0 {
		result.R2 = fptr(1 - residualSum/totalSum)
	}
	return result
}
func invert3(matrix [3][3]float64) ([3][3]float64, bool) {
	var augmented [3][6]float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			augmented[i][j] = matrix[i][j]
		}
		augmented[i][i+3] = 1
	}
	for i := 0; i < 3; i++ {
		pivot := i
		for j := i + 1; j < 3; j++ {
			if math.Abs(augmented[j][i]) > math.Abs(augmented[pivot][i]) {
				pivot = j
			}
		}
		if math.Abs(augmented[pivot][i]) < 1e-9 {
			return [3][3]float64{}, false
		}
		augmented[i], augmented[pivot] = augmented[pivot], augmented[i]
		value := augmented[i][i]
		for j := 0; j < 6; j++ {
			augmented[i][j] /= value
		}
		for k := 0; k < 3; k++ {
			if k == i {
				continue
			}
			factor := augmented[k][i]
			for j := 0; j < 6; j++ {
				augmented[k][j] -= factor * augmented[i][j]
			}
		}
	}
	var inverse [3][3]float64
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			inverse[i][j] = augmented[i][j+3]
		}
	}
	return inverse, true
}
