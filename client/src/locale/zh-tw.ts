export const logout = '登出'
export const unsupported = '此瀏覽器不支援 WebRTC'
export const admin_loggedin = '您已以管理員身份登入'
export const you = '您'
export const somebody = '某人'
export const send_a_message = '發送訊息'

export const side = {
  chat: '聊天',
  files: '檔案',
  settings: '設定',
  users: '使用者',
}

export const connect = {
  login_title: '請登入',
  invitation_title: '您已被邀請加入此房間',
  displayname: '輸入您的顯示名稱',
  password: '密碼',
  connect: '連接',
  error: '登入錯誤',
  empty_displayname: '顯示名稱不能為空。',
}

export const context = {
  ignore: '忽略',
  unignore: '取消忽略',
  mute: '靜音',
  unmute: '取消靜音',
  release: '強制釋放控制',
  take: '強制獲取控制',
  give: '給予控制',
  kick: '踢出',
  ban: '封鎖IP',
  confirm: {
    kick_title: '踢出 {name}？',
    kick_text: '您確定要踢出 {name} 嗎？',
    ban_title: '封鎖 {name}？',
    ban_text: '您要封鎖 {name} 嗎？您需要重啟伺服器才能撤销此操作。',
    mute_title: '靜音 {name}？',
    mute_text: '您確定要靜音 {name} 嗎？',
    unmute_title: '取消靜音 {name}？',
    unmute_text: '您要取消 {name} 的靜音嗎？',
    button_yes: '是',
    button_cancel: '取消',
  },
}

export const controls = {
  release: '釋放控制',
  request: '請求控制',
  lock: '鎖定控制',
  unlock: '解鎖控制',
  has: '您有控制權',
  hasnot: '您沒有控制權',
}

export const locks = {
  control: {
    lock: '鎖定控制（對使用者）',
    unlock: '解鎖控制（對使用者）',
    locked: '控制已鎖定（對使用者）',
    unlocked: '控制已解鎖（對使用者）',
    notif_locked: '鎖定了使用者控制',
    notif_unlocked: '解鎖了使用者控制',
  },
  login: {
    lock: '鎖定房間（對使用者）',
    unlock: '解鎖房間（對使用者）',
    locked: '房間已鎖定（對使用者）',
    unlocked: '房間已解鎖（對使用者）',
    notif_locked: '鎖定了房間',
    notif_unlocked: '解鎖了房間',
  },
  file_transfer: {
    lock: '鎖定檔案傳輸（對使用者）',
    unlock: '解鎖檔案傳輸（對使用者）',
    locked: '檔案傳輸已鎖定（對使用者）',
    unlocked: '檔案傳輸已解鎖（對使用者）',
    notif_locked: '鎖定了檔案傳輸',
    notif_unlocked: '解鎖了檔案傳輸',
  },
}

export const setting = {
  scroll: '捲動靈敏度',
  scroll_invert: '反轉捲動',
  autoplay: '自動播放影片',
  ignore_emotes: '忽略表情符號',
  chat_sound: '播放聊天聲音',
  keyboard_layout: '鍵盤配置',
  broadcast_title: '直播',
}

export const connection = {
  logged_out: '您已登出。',
  reconnecting: '重新連接中...',
  connected: '已連接',
  disconnected: '已斷開',
  kicked: '您已被移出此房間。',
  button_confirm: '確定',
}

export const notifications = {
  connected: '{name} 已連接',
  disconnected: '{name} 已斷開',
  controls_taken: '{name} 獲取了控制權',
  controls_taken_force: '強制獲取了控制權',
  controls_taken_steal: '從 {name} 處獲取了控制權',
  controls_released: '{name} 釋放了控制權',
  controls_released_force: '強制釋放了控制權',
  controls_released_steal: '釋放了 {name} 的控制權',
  controls_given: '將控制權給予了 {name}',
  controls_has: '{name} 擁有控制權',
  controls_has_alt: '但我已告知對方您想要控制權',
  controls_requesting: '{name} 正在請求控制權',
  resolution: '將解析度更改為 {width}x{height}@{rate}',
  banned: '封鎖了 {name}',
  kicked: '踢出了 {name}',
  muted: '靜音了 {name}',
  unmuted: '取消了 {name} 的靜音',
}

export const files = {
  downloads: '下載',
  uploads: '上傳',
  upload_here: '點擊或拖曳檔案到此處上傳',
}
