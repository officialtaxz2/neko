export const logout = '로그아웃'
export const unsupported = '이 웹 브라우저는 WebRTC를 지원하지 않습니다'
export const admin_loggedin = '관리자로 로그인되었습니다'
export const you = '나'
export const somebody = '누군가'
export const send_a_message = '메시지 보내기'

export const side = {
  chat: '채팅',
  files: '파일',
  settings: '설정',
  users: '사용자',
}

export const connect = {
  login_title: '로그인해 주세요',
  invitation_title: '이 방에 초대되었습니다',
  displayname: '표시 이름을 입력하세요',
  password: '비밀번호',
  connect: '연결',
  error: '로그인 오류',
  empty_displayname: '표시 이름은 비워둘 수 없습니다.',
}

export const context = {
  ignore: '무시',
  unignore: '무시 취소',
  mute: '음소거',
  unmute: '음소거 해제',
  release: '강제 컨트롤 해제',
  take: '강제 컨트롤 획득',
  give: '컨트롤 양도',
  kick: '추방',
  ban: 'IP 차단',
  confirm: {
    kick_title: '{name}님을 추방하시겠습니까?',
    kick_text: '{name}님을 추방하시겠습니까?',
    ban_title: '{name}님을 차단하시겠습니까?',
    ban_text: '{name}님을 차단하시겠습니까? 서버를 재시작해야 취소할 수 있습니다.',
    mute_title: '{name}님을 음소거하시겠습니까?',
    mute_text: '{name}님을 음소거하시겠습니까?',
    unmute_title: '{name}님의 음소거를 해제하시겠습니까?',
    unmute_text: '{name}님의 음소거를 해제하시겠습니까?',
    button_yes: '예',
    button_cancel: '취소',
  },
}

export const controls = {
  release: '컨트롤 해제',
  request: '컨트롤 요청',
  lock: '컨트롤 잠금',
  unlock: '컨트롤 잠금 해제',
  has: '컨트롤을 가지고 있습니다',
  hasnot: '컨트롤이 없습니다',
  mic_on: '마이크 활성화',
  mic_off: '마이크 비활성화',
  mic_error: '마이크 오류',
}

export const locks = {
  control: {
    lock: '컨트롤 잠금 (사용자용)',
    unlock: '컨트롤 잠금 해제 (사용자용)',
    locked: '컨트롤 잠김 (사용자용)',
    unlocked: '컨트롤 잠금 해제됨 (사용자용)',
    notif_locked: '사용자의 컨트롤을 잠갔습니다',
    notif_unlocked: '사용자의 컨트롤 잠금을 해제했습니다',
  },
  login: {
    lock: '방 잠금 (사용자용)',
    unlock: '방 잠금 해제 (사용자용)',
    locked: '방 잠김 (사용자용)',
    unlocked: '방 잠금 해제됨 (사용자용)',
    notif_locked: '방을 잠갔습니다',
    notif_unlocked: '방 잠금을 해제했습니다',
  },
  file_transfer: {
    lock: '파일 전송 잠금 (사용자용)',
    unlock: '파일 전송 잠금 해제 (사용자용)',
    locked: '파일 전송 잠김 (사용자용)',
    unlocked: '파일 전송 잠금 해제됨 (사용자용)',
    notif_locked: '파일 전송을 잠갔습니다',
    notif_unlocked: '파일 전송 잠금을 해제했습니다',
  },
}

export const setting = {
  scroll: '스크롤 감도',
  scroll_invert: '스크롤 반전',
  trackpad_mode: '트랙패드 모드',
  trackpad_mode_description: '터치패드처럼 커서를 상대적으로 이동',
  trackpad_mode_mobile_hint: '두 손가락으로 스크롤, 탭하여 클릭, 길게 눌러 우클릭',
  mobile: '모바일',
  desktop_only_inactive: '터치 전용',
  autoplay: '동영상 자동 재생',
  ignore_emotes: '이모지 무시',
  chat_sound: '채팅 소리 재생',
  keyboard_layout: '키보드 레이아웃',
  broadcast_title: '실시간 방송',
}

export const connection = {
  logged_out: '로그아웃 했습니다.',
  reconnecting: '다시 접속하는 중...',
  connected: '연결됨',
  disconnected: '연결 해제됨',
  kicked: '이 방에서 추방됐습니다.',
  button_confirm: '확인',
}

export const notifications = {
  connected: '{name} 님이 접속하셨습니다',
  disconnected: '{name} 님이 접속을 끊었습니다',
  controls_taken: '{name} 님이 컨트롤을 가져갔습니다',
  controls_taken_force: '강제로 컨트롤을 가져갔습니다',
  controls_taken_steal: '{name} 님으로부터 컨트롤을 가져갔습니다',
  controls_released: '{name} 님이 컨트롤을 해제했습니다',
  controls_released_force: '강제로 컨트롤을 해제했습니다',
  controls_released_steal: '{name} 님의 컨트롤을 해제했습니다',
  controls_given: '{name} 님에게 컨트롤을 넘겼습니다',
  controls_has: '{name} 님이 컨트롤을 가지고 있습니다',
  controls_has_alt: '하지만 그 사람에게 원한다고 알려줬습니다',
  controls_requesting: '{name} 님이 컨트롤을 요청하고 있습니다',
  resolution: '해상도를 {width}x{height}@{rate}로 변경했습니다',
  banned: '{name} 님을 차단했습니다',
  kicked: '{name} 님을 추방했습니다',
  muted: '{name} 님을 음소거했습니다',
  unmuted: '{name} 님의 음소거를 해제했습니다',
}

export const files = {
  downloads: '다운로드',
  uploads: '업로드',
  upload_here: '여기를 클릭하거나 파일을 드래그하여 업로드',
}
