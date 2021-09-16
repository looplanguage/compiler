package code

import "testing"

func TestMake(t *testing.T) {
	tests := []struct {
		op       OpCode
		operands []int
		expected []byte
	}{
		{OpConstant, []int{65534}, []byte{byte(OpConstant), 255, 254}},
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
