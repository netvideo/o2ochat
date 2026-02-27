# O2OChat

🌍 **[English](README_EN.md)** | **[中文](README.md)** | **[Español](README_ES.md)** | **[Français](README_FR.md)** | **[Deutsch](README_DE.md)** | **[日本語](README_JA.md)** | **[한국어](README_KO.md)** | **[Русский](README_RU.md)** | **[العربية](README_AR.md)** | **[עברית](README_HE.md)** | **[Bahasa Melayu](README_MS.md)**

## 순수 P2P 인스턴트 메시징 소프트웨어

O2OChat은 중앙 서버에 메시지를 저장하지 않는 순수 P2P(피어 투 피어) 인스턴트 메시징 소프트웨어입니다. 모든 통신은 사용자 간에 직접 이루어집니다.

### 핵심 기능

- 🔒 **종단 간 암호화** - 모든 메시지는 AES-256-GCM 암호화 사용
- 🌐 **순수 P2P 아키텍처** - 중앙 서버 없음, 직접 통신
- 📱 **멀티 플랫폼 지원** - Android, iOS, Windows, Linux, macOS, HarmonyOS
- 📁 **파일 전송** - 중단된 전송 재개, 다중 소스 다운로드, Merkle 트리 검증
- 🌍 **16개 언어** - 중국어, 영어, 일본어, 한국어, 독일어, 프랑스어, 스페인어, 러시아어, 말레이어, 히브리어, 아랍어, 티베트어, 몽골어, 위구르어, 중국어(번체)

### 빠른 시작

```bash
# 프로젝트 복제
git clone https://github.com/yourusername/o2ochat.git
cd o2ochat

# 빌드
go build -o o2ochat ./cmd/o2ochat

# 실행
./o2ochat
```

### 프로젝트 구조

```
o2ochat/
├── cmd/              # 진입점
├── pkg/              # 핵심 라이브러리
│   ├── identity/     # 신원 관리
│   ├── transport/    # 네트워크 전송
│   ├── signaling/    # 시그널링 서비스
│   ├── crypto/       # 암호화 모듈
│   ├── storage/      # 데이터 저장
│   ├── filetransfer/ # 파일 전송
│   └── media/        # 오디오/비디오 처리
├── ui/               # 사용자 인터페이스
├── cli/              # 명령줄 도구
├── tests/            # 테스트
├── docs/             # 문서
└── scripts/          # 빌드 스크립트
```

### 기술 스택

- **Go 1.21+** - 백엔드 코어
- **Protocol Buffers** - 직렬화
- **QUIC/WebRTC** - P2P 전송
- **SQLite** - 로컬 저장소
- **Fyne** - 데스크톱 GUI
- **Jetpack Compose** - Android UI
- **SwiftUI** - iOS UI
- **ArkTS** - HarmonyOS UI

### 기여하기

기여를 환영합니다! [Contributing Guide](CONTRIBUTING.md)를 읽어주세요.

### 라이선스

MIT License - 자세한 내용은 [LICENSE](LICENSE) 파일을 참조하세요.

### 연락처

- 프로젝트 홈페이지: https://o2ochat.io
- 이슈 트래커: https://github.com/yourusername/o2ochat/issues
- 이메일: support@o2ochat.io

---

<p align="center">
  <b>순수 P2P | 종단 간 암호화 | 자유로운 커뮤니케이션</b>
</p>
