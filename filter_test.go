package suzuitoql

/*
func TestValidateFilter(t *testing.T) {
	testCases := []struct {
		desc        string
		input       interface{}
		expectedErr string
	}{
		{
			desc: "",
			input: func(topic *entities.Topic) bool {
				return false
			},
		},
		{
			desc: "Invalid function: Empty arguments",
			input: func() bool {
				return false
			},
			expectedErr: "ErrInvalidFilter: Number of arguments must be larger than 1",
		},
		{
			desc: "Invalid function: Arg0 is not topic",
			input: func(a int) bool {
				return false
			},
			expectedErr: "ErrInvalidFilter: Arg0 is not *entities.Topic",
		},
		{
			desc: "Number of return values must be 1",
			input: func(topic *entities.Topic) {
			},
			expectedErr: "ErrInvalidFilter: Number of return values must be 1",
		},
		{
			desc: "Return value must be bool",
			input: func(topic *entities.Topic) int {
				return 1
			},
			expectedErr: "ErrInvalidFilter: Return value must be bool",
		},
		{
			desc:        "Not function",
			input:       1,
			expectedErr: "ErrInvalidFilter: Not function",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			err := validateFilter(tC.input)
			if err != nil {
				assert.Equal(t, tC.expectedErr, err.Error())
				return
			}
		})
	}
}

func TestFilter(t *testing.T) {
	testCases := []struct {
		desc            string
		inputExpression string
		inputTopic      entities.Topic
		expected        bool
		expectedErr     string
	}{
		{
			desc: "",
			inputExpression: `
			mustTrueWithArg1("A")
			`,
			expected: true,
		},
		{
			desc: "",
			inputExpression: `
			mustTrueWithArg1("A") || mustFalseWithArg1("B")
			`,
			expected: true,
		},
		{
			desc: "",
			inputExpression: `
			mustTrueWithArg1("A") && mustFalseWithArg1("B")
			`,
			expected: false,
		},
		{
			desc: "",
			inputExpression: `
			mustTrueWithArg1("A") && mustFalseWithArg1("B") || mustTrueWithArg1("C")
			`,
			expected: true,
		},
		{
			desc: "",
			inputExpression: `
			mustTrueWithArg1("A") && (mustFalseWithArg1("B") || mustFalseWithArg1("C"))
			`,
			expected: false,
		},
		{
			desc: "",
			inputExpression: `
			(mustTrueWithArg1("A") || mustFalseWithArg1("B"))
			&&
			(mustFalseWithArg1("C") || mustFalseWithArg1("C"))
			`,
			expected: false,
		},
		{
			desc: "",
			inputExpression: `
			(
			    (mustTrueWithArg1("A") || mustTrueWithArg1("B"))
			    &&
			    (mustTrueWithArg1("C") || mustTrueWithArg1("C"))
			)
			||
			(
			    (mustFalseWithArg1("A") || mustFalseWithArg1("B"))
			    &&
			    (mustFalseWithArg1("C") || mustFalseWithArg1("C"))
			)
			`,
			expected: true,
		},
		{
			desc: "",
			inputExpression: `
			mustTrueWithIntArg1(1)
			`,
			expected: true,
		},
		{
			desc: "",
			inputExpression: `
			mustTrueWithBoolArg1(true)
			`,
			expected: true,
		},
		{
			desc: "",
			inputExpression: `
			mustTrueWithFloatArg1(1.1)
			`,
			expected: true,
		},
		{
			desc: "",
			inputExpression: `
			mustTrueWithoutArg()
			`,
			expected: true,
		},
		// 論理式の文法エラー
		// TODO: エラーメッセージが意味不明（Go言語のシンタックスエラーがそのまま返ってくる）
		//       式のどの部分がおかしいか？わかるとベストなんだが・・・
		{
			desc:            "Sytax error: empty string",
			inputExpression: ``,
			expectedErr:     "1:1: expected operand, found 'EOF'",
		},
		{
			desc:            "Sytax error: empty string",
			inputExpression: `mustTrueWithArg1("A") &&`,
			expectedErr:     "3:1: expected operand, found '}'",
		},
		// 評価時のエラー
		{
			desc: "Evaluation error: Call with too few input arguments",
			inputExpression: `
			mustTrueWithArg1()
			`,
			expectedErr: "reflect: Call with too few input arguments",
		},
		{
			desc: "Evaluation error: Call with too many input arguments",
			inputExpression: `
			mustTrueWithArg1("a", "b")
			`,
			expectedErr: "reflect: Call with too many input arguments",
		},
		{
			desc: "Evaluation error: Call using int as type string",
			inputExpression: `
			mustTrueWithArg1(1)
			`,
			expectedErr: "reflect: Call using int as type string",
		},
		{
			desc: "Evaluation error: Function not found",
			inputExpression: `
			dummyFunc("a")
			`,
			expectedErr: "Function not found: dummyFunc",
		},
		{
			desc: "Unsupported op",
			inputExpression: `
			mustTrueWithArg1("A") + mustTrueWithArg1("B")
			`,
			expectedErr: "Unsupported op: +",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			fg, err := GenerateFilterFromString(tC.inputExpression)
			if err != nil {
				assert.Equal(t, tC.expectedErr, err.Error())
				return
			}
			real, err := fg.Eval(&tC.inputTopic)
			if err != nil {
				assert.Equal(t, tC.expectedErr, err.Error())
				return
			}
			assert.Equal(t, tC.expected, real)
		})
	}
}

func TestPerformance(t *testing.T) {
	testCases := []struct {
		desc                      string
		inputExpressions          []string
		inputTopic                entities.Topic
		expectedLimitMilliSeconds int
		expected                  bool
	}{
		{
			desc: "1000個の句を繋げた、1個のフィルタの評価が0.010秒以内に終わる",
			inputExpressions: func() []string {
				ret := []string{}
				for i := 0; i < 1; i++ {
					each := []string{}
					for j := 0; j < 1000; j++ {
						each = append(each, `mustTrueWithArg1("A")`)
					}
					ret = append(ret, strings.Join(each, "||"))
				}
				return ret
			}(),
			expectedLimitMilliSeconds: 10, // 0.010 seconds
			expected:                  true,
		},
		// {
		// 	desc: "1000個の句を繋げた、1000個のフィルタの評価が10秒以内に終わる（このテスト長い）",
		// 	inputExpressions: func() []string {
		// 		ret := []string{}
		// 		for i := 0; i < 1000; i++ {
		// 			each := []string{}
		// 			for j := 0; j < 1000; j++ {
		// 				each = append(each, `mustTrueWithArg1("A")`)
		// 			}
		// 			ret = append(ret, strings.Join(each, "||"))
		// 		}
		// 		return ret
		// 	}(),
		// 	expectedLimitMilliSeconds: 10000, // 10 seconds
		// 	expected:                  true,
		// },
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			start := time.Now()
			defer func() {
				end := time.Now()
				diff := end.UnixNano() - start.UnixNano()
				assert.LessOrEqual(
					t,
					diff,
					int64(tC.expectedLimitMilliSeconds)*int64(10e+6),
				)
			}()
			fgs := []FilterGroup{}
			for _, e := range tC.inputExpressions {
				fg, err := Parse(e)
				if err != nil {
					assert.Fail(t, err.Error())
					return
				}
				fgs = append(fgs, fg)
			}
			for _, fg := range fgs {
				real, err := fg.Eval(&tC.inputTopic)
				if err != nil {
					assert.Fail(t, err.Error())
					return
				}
				assert.Equal(t, tC.expected, real)
			}
		})
	}
}

func TestFilterFunctionValidation(t *testing.T) {
	err := ValidateFilters()
	assert.Nil(t, err)
}
*/
