export const logout = 'ログアウト'
export const unsupported = 'このウェブブラウザはWebRTCをサポートしていません'
export const admin_loggedin = '管理者としてログインしています'
export const you = 'あなた'
export const somebody = '誰か'
export const send_a_message = 'メッセージを送信'

export const side = {
  chat: 'チャット',
  files: 'ファイル',
  settings: '設定',
  users: 'ユーザー',
}

export const connect = {
  login_title: 'ログインしてください',
  invitation_title: 'この部屋に招待されました',
  displayname: '表示名を入力してください',
  password: 'パスワード',
  connect: '接続',
  error: 'ログインエラー',
  empty_displayname: '表示名を空にすることはできません。',
}

export const context = {
  ignore: '無視',
  unignore: '無視をやめる',
  mute: 'ミュート',
  unmute: 'ミュート解除',
  release: '強制的にコントロールを解放',
  take: '強制的にコントロールを取得',
  give: 'コントロールを渡す',
  kick: 'キック',
  ban: 'IPをBAN',
  confirm: {
    kick_title: '{name}をキックしますか？',
    kick_text: '{name}をキックしてもよいですか？',
    ban_title: '{name}をBANしますか？',
    ban_text: '{name}をBANしますか？サーバーを再起動しないと元に戻せません。',
    mute_title: '{name}をミュートしますか？',
    mute_text: '{name}をミュートしてもよいですか？',
    unmute_title: '{name}のミュートを解除しますか？',
    unmute_text: '{name}のミュートを解除しますか？',
    button_yes: 'はい',
    button_cancel: 'キャンセル',
  },
}

export const controls = {
  release: 'コントロールを解放',
  request: 'コントロールをリクエスト',
  lock: 'コントロールをロック',
  unlock: 'コントロールのロックを解除',
  has: 'あなたはコントロールを持っています',
  hasnot: 'あなたはコントロールを持っていません',
  mic_on: 'マイクを有効にする',
  mic_off: 'マイクを無効にする',
  mic_error: 'マイクエラー',
}

export const locks = {
  control: {
    lock: 'コントロールをロック（ユーザー向け）',
    unlock: 'コントロールのロックを解除（ユーザー向け）',
    locked: 'コントロールがロックされています（ユーザー向け）',
    unlocked: 'コントロールのロックが解除されました（ユーザー向け）',
    notif_locked: 'ユーザーのコントロールをロックしました',
    notif_unlocked: 'ユーザーのコントロールのロックを解除しました',
  },
  login: {
    lock: '部屋をロック（ユーザー向け）',
    unlock: '部屋のロックを解除（ユーザー向け）',
    locked: '部屋がロックされています（ユーザー向け）',
    unlocked: '部屋のロックが解除されました（ユーザー向け）',
    notif_locked: '部屋をロックしました',
    notif_unlocked: '部屋のロックを解除しました',
  },
  file_transfer: {
    lock: 'ファイル転送をロック（ユーザー向け）',
    unlock: 'ファイル転送のロックを解除（ユーザー向け）',
    locked: 'ファイル転送がロックされています（ユーザー向け）',
    unlocked: 'ファイル転送のロックが解除されました（ユーザー向け）',
    notif_locked: 'ファイル転送をロックしました',
    notif_unlocked: 'ファイル転送のロックを解除しました',
  },
}

export const setting = {
  scroll: 'スクロールの感度',
  scroll_invert: 'スクロールを反転する',
  trackpad_mode: 'トラックパッドモード',
  trackpad_mode_description: 'タッチパッドのようにカーソルを相対的に移動',
  trackpad_mode_mobile_hint: '2本指でスクロール、タップでクリック、長押しで右クリック',
  mobile: 'モバイル',
  desktop_only_inactive: 'タッチのみ',
  autoplay: '動画を自動再生する',
  ignore_emotes: '絵文字を無視する',
  chat_sound: 'チャットで音を再生する',
  keyboard_layout: 'キーボード配列',
  broadcast_title: 'ライブ配信',
}

export const connection = {
  logged_out: 'ログアウトしました',
  reconnecting: '再接続中...',
  connected: '接続しました',
  disconnected: '切断しました',
  kicked: 'あなたはこの部屋から追い出されました',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} が接続しました',
  disconnected: '{name} が切断しました',
  controls_taken: '{name} がコントロールを取得しました',
  controls_taken_force: 'コントロールを強制的に取得しました',
  controls_taken_steal: '{name} からコントロールを取得しました',
  controls_released: '{name} がコントロールを解放しました',
  controls_released_force: 'コントロールを強制的に解放しました',
  controls_released_steal: '{name} のコントロールを解放しました',
  controls_given: '{name} にコントロールを渡しました',
  controls_has: '{name} がコントロールを持っています',
  controls_has_alt: 'しかし、その人にあなたが欲しいと伝えました',
  controls_requesting: '{name} がコントロールをリクエストしています',
  resolution: '解像度を {width}x{height}@{rate} に変更しました',
  banned: '{name} をBANしました',
  kicked: '{name} をキックしました',
  muted: '{name} をミュートしました',
  unmuted: '{name} のミュートを解除しました',
}

export const files = {
  downloads: 'ダウンロード',
  uploads: 'アップロード',
  upload_here: 'ここをクリックまたはファイルをドラッグしてアップロード',
}
