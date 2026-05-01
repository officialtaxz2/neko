export const logout = 'Выйти'
export const unsupported = 'Этот браузер не поддерживает WebRTC'
export const admin_loggedin = 'Вы вошли как администратор'
export const you = 'Вы'
export const somebody = 'Кто-то'
export const send_a_message = 'Отправить сообщение'

export const side = {
  chat: 'Чат',
  files: 'Файлы',
  settings: 'Настройки',
  users: 'Пользователи',
}

export const connect = {
  login_title: 'Пожалуйста, войдите',
  invitation_title: 'Вас пригласили в эту комнату',
  displayname: 'Введите ваше отображаемое имя',
  password: 'Пароль',
  connect: 'Подключиться',
  error: 'Ошибка входа',
  empty_displayname: 'Отображаемое имя не может быть пустым.',
}

export const context = {
  ignore: 'Игнорировать',
  unignore: 'Перестать игнорировать',
  mute: 'Заглушить',
  unmute: 'Включить звук',
  release: 'Принудительно освободить управление',
  take: 'Принудительно взять управление',
  give: 'Передать управление',
  kick: 'Кикнуть',
  ban: 'Забанить IP',
  confirm: {
    kick_title: 'Кикнуть {name}?',
    kick_text: 'Вы уверены, что хотите кикнуть {name}?',
    ban_title: 'Забанить {name}?',
    ban_text: 'Вы хотите забанить {name}? Для отмены потребуется перезапустить сервер.',
    mute_title: 'Заглушить {name}?',
    mute_text: 'Вы уверены, что хотите заглушить {name}?',
    unmute_title: 'Включить звук {name}?',
    unmute_text: 'Вы хотите включить звук {name}?',
    button_yes: 'Да',
    button_cancel: 'Отмена',
  },
}

export const controls = {
  release: 'Освободить управление',
  request: 'Запросить управление',
  lock: 'Заблокировать управление',
  unlock: 'Разблокировать управление',
  has: 'У вас есть управление',
  hasnot: 'У вас нет управления',
  mic_on: 'Включить микрофон',
  mic_off: 'Выключить микрофон',
  mic_error: 'Ошибка микрофона',
}

export const locks = {
  control: {
    lock: 'Заблокировать управление (для пользователей)',
    unlock: 'Разблокировать управление (для пользователей)',
    locked: 'Управление заблокировано (для пользователей)',
    unlocked: 'Управление разблокировано (для пользователей)',
    notif_locked: 'заблокировал управление для пользователей',
    notif_unlocked: 'разблокировал управление для пользователей',
  },
  login: {
    lock: 'Заблокировать комнату (для пользователей)',
    unlock: 'Разблокировать комнату (для пользователей)',
    locked: 'Комната заблокирована (для пользователей)',
    unlocked: 'Комната разблокирована (для пользователей)',
    notif_locked: 'заблокировал комнату',
    notif_unlocked: 'разблокировал комнату',
  },
  file_transfer: {
    lock: 'Заблокировать передачу файлов (для пользователей)',
    unlock: 'Разблокировать передачу файлов (для пользователей)',
    locked: 'Передача файлов заблокирована (для пользователей)',
    unlocked: 'Передача файлов разблокирована (для пользователей)',
    notif_locked: 'заблокировал передачу файлов',
    notif_unlocked: 'разблокировал передачу файлов',
  },
}

export const setting = {
  scroll: 'Чувствительность прокрутки',
  scroll_invert: 'Инвертировать прокрутку',
  trackpad_mode: 'Режим тачпада',
  trackpad_mode_description: 'Перемещать курсор относительно, как тачпад',
  trackpad_mode_mobile_hint: 'Два пальца для прокрутки, касание для клика, удержание для правого клика',
  mobile: 'Мобильный',
  desktop_only_inactive: 'Только сенсор',
  autoplay: 'Автовоспроизведение видео',
  ignore_emotes: 'Игнорировать эмодзи',
  chat_sound: 'Звук чата',
  keyboard_layout: 'Раскладка клавиатуры',
  broadcast_title: 'Прямая трансляция',
}

export const connection = {
  logged_out: 'Вы вышли из системы.',
  reconnecting: 'Переподключение...',
  connected: 'Подключено',
  disconnected: 'Отключено',
  kicked: 'Вас выгнали из этой комнаты.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} подключился',
  disconnected: '{name} отключился',
  controls_taken: '{name} взял управление',
  controls_taken_force: 'принудительно взял управление',
  controls_taken_steal: 'взял управление у {name}',
  controls_released: '{name} освободил управление',
  controls_released_force: 'принудительно освободил управление',
  controls_released_steal: 'освободил управление у {name}',
  controls_given: 'передал управление {name}',
  controls_has: '{name} имеет управление',
  controls_has_alt: 'Но я сообщил человеку, что вы хотели его',
  controls_requesting: '{name} запрашивает управление',
  resolution: 'изменил разрешение на {width}x{height}@{rate}',
  banned: 'забанил {name}',
  kicked: 'кикнул {name}',
  muted: 'заглушил {name}',
  unmuted: 'включил звук {name}',
}

export const files = {
  downloads: 'Загрузки',
  uploads: 'Выгрузки',
  upload_here: 'Нажмите или перетащите файлы сюда для загрузки',
}
