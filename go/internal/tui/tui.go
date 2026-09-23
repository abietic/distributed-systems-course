// Package tui 提供中英混排的终端对齐工具，供各个 Lab 共用。
package tui

import (
	"fmt"
	"strings"
)

// Dispw 计算字符串在等宽终端里占几列（CJK 全角字符占 2 列）。
// Go 的 %-20s 是按 rune 计数的，中文表格会错位，所以自己算。
func Dispw(s string) int {
	w := 0
	for _, r := range s {
		if r >= 0x1100 && (r <= 0x115F ||
			(r >= 0x2E80 && r <= 0xA4CF && r != 0x303F) ||
			(r >= 0xAC00 && r <= 0xD7A3) ||
			(r >= 0xF900 && r <= 0xFAFF) ||
			(r >= 0xFE30 && r <= 0xFE6F) ||
			(r >= 0xFF00 && r <= 0xFF60) ||
			(r >= 0xFFE0 && r <= 0xFFE6)) {
			w += 2
		} else {
			w++
		}
	}
	return w
}

func PadR(s string, w int) string {
	if d := w - Dispw(s); d > 0 {
		return s + strings.Repeat(" ", d)
	}
	return s
}

func PadL(s string, w int) string {
	if d := w - Dispw(s); d > 0 {
		return strings.Repeat(" ", d) + s
	}
	return s
}

// Head 打印一个实验的标题块。
func Head(part, n int, title, subtitle string) {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Printf("  Lab %d-%d · %s\n  %s\n", part, n, title, subtitle)
	fmt.Printf("%s\n", strings.Repeat("═", 78))
}

// Row 打印一行「键 + 值」。
func Row(k string, v ...any) {
	fmt.Printf("    %s %s\n", PadR(k, 30), fmt.Sprint(v...))
}

// TableHead / TableRow 按显示宽度对齐：第一列左对齐，其余右对齐。
func TableHead(cols []string, w []int) { TableHeadA(cols, w, defaultAlign(len(cols))) }
func TableRow(cells []string, w []int) { TableRowA(cells, w, defaultAlign(len(cells))) }

func defaultAlign(n int) string {
	if n == 0 {
		return ""
	}
	return "L" + strings.Repeat("R", n-1)
}

// TableHeadA / TableRowA 支持逐列指定对齐，align 形如 "LRRL"（L 左对齐，R 右对齐）。
func TableHeadA(cols []string, w []int, align string) {
	fmt.Println(layout(cols, w, align))
	t := 0
	for _, x := range w {
		t += x
	}
	fmt.Printf("    %s\n", strings.Repeat("─", t))
}

func TableRowA(cells []string, w []int, align string) { fmt.Println(layout(cells, w, align)) }

func layout(cells []string, w []int, align string) string {
	line := "    "
	for i, c := range cells {
		a := byte('R')
		if i < len(align) {
			a = align[i]
		}
		if a == 'L' {
			line += PadR(c, w[i])
		} else {
			line += PadL(c, w[i])
		}
	}
	return strings.TrimRight(line, " ")
}

func Bar(n int) string { return strings.Repeat("═", n) }

// HeadN 与 Head 相同，但实验编号由调用方自己给（比如 "3A-1"）。
func HeadN(label, title, subtitle string) {
	fmt.Printf("\n%s\n", strings.Repeat("═", 78))
	fmt.Printf("  Lab %s · %s\n  %s\n", label, title, subtitle)
	fmt.Printf("%s\n", strings.Repeat("═", 78))
}
