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
  login_title: 'Войдите',
  invitation_title: 'Вас пригласили в эту комнату',
  displayname: 'Введите своё имя',
  password: 'Пароль',
  connect: 'Подключиться',
  error: 'Ошибка входа',
  empty_displayname: 'Имя не может быть пустым.',
}

export const context = {
  ignore: 'Игнорировать',
  unignore: 'Не игнорировать',
  mute: 'Заглушить',
  unmute: 'Включить звук',
  release: 'Принудительно снять управление',
  take: 'Принудительно взять управление',
  give: 'Передать управление',
  kick: 'Выгнать',
  ban: 'Заблокировать IP',
  confirm: {
    kick_title: 'Выгнать {name}?',
    kick_text: 'Вы уверены, что хотите выгнать {name}?',
    ban_title: 'Заблокировать {name}?',
    ban_text: 'Вы хотите заблокировать {name}? Чтобы отменить, вам нужно перезапустить сервер.',
    mute_title: 'Заглушить {name}?',
    mute_text: 'Вы уверены, что хотите заглушить {name}?',
    unmute_title: 'Включить звук {name}?',
    unmute_text: 'Вы хотите включить звук {name}?',
    button_yes: 'Да',
    button_cancel: 'Отмена',
  },
}

export const controls = {
  release: 'Снять управление',
  request: 'Запросить управление',
  lock: 'Заблокировать управление',
  unlock: 'Разблокировать управление',
  has: 'У вас есть управление',
  hasnot: 'У вас нет управления',
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
  controls_taken_steal: 'забрал управление у {name}',
  controls_released: '{name} снял управление',
  controls_released_force: 'принудительно снял управление',
  controls_released_steal: 'снял управление у {name}',
  controls_given: 'передал управление {name}',
  controls_has: '{name} имеет управление',
  controls_has_alt: 'Но я сообщил этому человеку, что вы этого хотели',
  controls_requesting: '{name} запрашивает управление',
  resolution: 'изменил разрешение на {width}x{height}@{rate}',
  banned: 'заблокировал {name}',
  kicked: 'выгнал {name}',
  muted: 'заглушил {name}',
  unmuted: 'включил звук {name}',
}

export const files = {
  downloads: 'Загрузки',
  uploads: 'Выгрузки',
  upload_here: 'Нажмите или перетащите файлы сюда для загрузки',
}
