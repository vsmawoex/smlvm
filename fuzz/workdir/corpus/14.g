func printStr(s string) {
	n:=len(s); for i:=0;i<n;i++ { printChar(s[i]) }
}
func main() { var a []int8; b:="hello"; a=b; printStr(a) }