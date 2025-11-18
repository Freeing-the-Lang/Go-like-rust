#!/usr/bin/env python3
import sys
import re

"""
Go → SpongeLang Meaning IR

초기 버전:
- println / fmt.Println 탐지
- 문자열 추출
- 기본적인 의미(IR) 생성
"""

def extract_prints(go_code: str):
    prints = []
    # 예: fmt.Println("Hello")
    pattern = r'Println\((.*?)\)'
    matches = re.findall(pattern, go_code)
    for m in matches:
        s = m.strip()
        # 문자열만 추출
        if s.startswith('"') and s.endswith('"'):
            prints.append(s.strip('"'))
    return prints


def main():
    if len(sys.argv) < 2:
        print("Usage: go_to_sponge_meaning.py <go_file>", file=sys.stderr)
        sys.exit(1)

    path = sys.argv[1]
    try:
        with open(path, "r", encoding="utf-8") as f:
            go_src = f.read()
    except:
        print("Error: cannot read Go file", file=sys.stderr)
        sys.exit(1)

    prints = extract_prints(go_src)

    # Meaning IR 출력
    print("program:")
    for p in prints:
        print(f"  print \"{p}\"")

    # Go 버전 fallback에서는 Enter 대기 있으므로 의미도 반영
    print("  wait-input")


if __name__ == "__main__":
    main()
