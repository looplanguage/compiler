package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       OpCode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
		{OpAdd, []int{}, []byte{byte(OpAdd)}},
		{OpMultiply, []int{}, []byte{byte(OpMultiply)}},
		{OpSubtract, []int{}, []byte{byte(OpSubtract)}},
		{OpGetLocal, []int{255}, []byte{byte(OpGetLocal), 255}},
		{OpClosure, []int{65534, 255}, []byte{byte(OpClosure), 255, 254, 255}},
	}

	for _, tc := range tests {
		instruction := Make(tc.op, tc.operands...)

		if len(instruction) != len(tc.expected) {
			t.Errorf("instruction has wrong length. want=%d. got=%d", len(tc.expected), len(instruction))
		}

		for i, b := range tc.expected {
			if instruction[i] != tc.expected[i] {
				t.Errorf("wrong byte at pos %d. want=%d, got=%d", i, b, instruction[i])
			}
		}
	}
}

func TestInstructionFormatting(t *testing.T) {
	instructions := []Instructions{
		Make(OpAdd),
		Make(OpConstant, 2),
		Make(OpConstant, 65535),
		Make(OpMultiply),
		Make(OpDivide),
		Make(OpSubtract),
		Make(OpEquals),
		Make(OpNotEquals),
		Make(OpGreaterThan),
		Make(OpNull),
		Make(OpGetLocal, 1),
		Make(OpClosure, 65535, 255),
	}

	expected := `[0000] OpAdd
[0001] OpConstant 2
[0004] OpConstant 65535
[0007] OpMultiply
[0008] OpDivide
[0009] OpSubtract
[0010] OpEquals
[0011] OpNotEquals
[0012] OpGreaterThan
[0013] OpNull
[0014] OpGetLocal 1
[0016] OpClosure 65535 255
`

	concatted := Instructions{}

	for _, ins := range instructions {
		concatted = append(concatted, ins...)
	}

	if concatted.String() != expected {
		t.Errorf("instructions wrongly formatted. \ngot=%q\nexpected=%q\n", concatted.String(), expected)
	}
}

func TestReadOperands(t *testing.T) {
	tests := []struct {
		op        OpCode
		operands  []int
		bytesRead int
	}{
		{OpConstant, []int{65535}, 2},
		{OpGetLocal, []int{255}, 1},
		{OpClosure, []int{65535, 255}, 3},
	}

	for _, tc := range tests {
		instruction := Make(tc.op, tc.operands...)

		def, err := Lookup(byte(tc.op))
		if err != nil {
			t.Fatalf("definition not found: %q\n", err)
		}

		operandsRead, n := ReadOperands(def, instruction[1:])
		if n != tc.bytesRead {
			t.Fatalf("wrong amount operands. want=%d. got=%d", tc.bytesRead, n)
		}

		for i, want := range tc.operands {
			if operandsRead[i] != want {
				t.Errorf("unexpected operand. want=%d. got=%d", want, operandsRead[i])
			}
		}
	}
}
