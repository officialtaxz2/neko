export const logout = 'Cerrar sesión'
export const unsupported = 'Este navegador no es compatible con WebRTC'
export const admin_loggedin = 'Has iniciado sesión como administrador'
export const you = 'Tú'
export const somebody = 'Alguien'
export const send_a_message = 'Enviar un mensaje'

export const side = {
  chat: 'Chat',
  files: 'Archivos',
  settings: 'Configuración',
  users: 'Usuarios',
}

export const connect = {
  login_title: 'Por favor inicia sesión',
  invitation_title: 'Has sido invitado a esta sala',
  displayname: 'Ingresa tu nombre para mostrar',
  password: 'Contraseña',
  connect: 'Conectar',
  error: 'Error de inicio de sesión',
  empty_displayname: 'El nombre para mostrar no puede estar vacío.',
}

export const context = {
  ignore: 'Ignorar',
  unignore: 'Dejar de ignorar',
  mute: 'Silenciar',
  unmute: 'Activar sonido',
  release: 'Forzar liberación de controles',
  take: 'Forzar toma de controles',
  give: 'Dar controles',
  kick: 'Expulsar',
  ban: 'Banear IP',
  confirm: {
    kick_title: '¿Expulsar a {name}?',
    kick_text: '¿Estás seguro de que quieres expulsar a {name}?',
    ban_title: '¿Banear a {name}?',
    ban_text: '¿Quieres banear a {name}? Necesitarás reiniciar el servidor para deshacer esto.',
    mute_title: '¿Silenciar a {name}?',
    mute_text: '¿Estás seguro de que quieres silenciar a {name}?',
    unmute_title: '¿Activar sonido de {name}?',
    unmute_text: '¿Quieres activar el sonido de {name}?',
    button_yes: 'Sí',
    button_cancel: 'Cancelar',
  },
}

export const controls = {
  release: 'Liberar controles',
  request: 'Solicitar controles',
  lock: 'Bloquear controles',
  unlock: 'Desbloquear controles',
  has: 'Tienes el control',
  hasnot: 'No tienes el control',
  mic_on: 'Activar micrófono',
  mic_off: 'Desactivar micrófono',
  mic_error: 'Error de micrófono',
}

export const locks = {
  control: {
    lock: 'Bloquear controles (para usuarios)',
    unlock: 'Desbloquear controles (para usuarios)',
    locked: 'Controles bloqueados (para usuarios)',
    unlocked: 'Controles desbloqueados (para usuarios)',
    notif_locked: 'bloqueó los controles para usuarios',
    notif_unlocked: 'desbloqueó los controles para usuarios',
  },
  login: {
    lock: 'Bloquear sala (para usuarios)',
    unlock: 'Desbloquear sala (para usuarios)',
    locked: 'Sala bloqueada (para usuarios)',
    unlocked: 'Sala desbloqueada (para usuarios)',
    notif_locked: 'bloqueó la sala',
    notif_unlocked: 'desbloqueó la sala',
  },
  file_transfer: {
    lock: 'Bloquear transferencia de archivos (para usuarios)',
    unlock: 'Desbloquear transferencia de archivos (para usuarios)',
    locked: 'Transferencia de archivos bloqueada (para usuarios)',
    unlocked: 'Transferencia de archivos desbloqueada (para usuarios)',
    notif_locked: 'bloqueó la transferencia de archivos',
    notif_unlocked: 'desbloqueó la transferencia de archivos',
  },
}

export const setting = {
  scroll: 'Sensibilidad del scroll',
  scroll_invert: 'Invertir scroll',
  trackpad_mode: 'Modo trackpad',
  trackpad_mode_description: 'Mover el cursor de forma relativa como un touchpad',
  trackpad_mode_mobile_hint: 'Dos dedos para desplazar, toca para hacer clic, mantén para clic derecho',
  mobile: 'Móvil',
  desktop_only_inactive: 'Solo táctil',
  autoplay: 'Reproducir video automáticamente',
  ignore_emotes: 'Ignorar emoticonos',
  chat_sound: 'Reproducir sonido del chat',
  keyboard_layout: 'Diseño del teclado',
  broadcast_title: 'Transmisión en vivo',
}

export const connection = {
  logged_out: 'Has sido desconectado.',
  reconnecting: 'Reconectando...',
  connected: 'Conectado',
  disconnected: 'Desconectado',
  kicked: 'Has sido expulsado de esta sala.',
  button_confirm: 'OK',
}

export const notifications = {
  connected: '{name} se conectó',
  disconnected: '{name} se desconectó',
  controls_taken: '{name} tomó los controles',
  controls_taken_force: 'tomó los controles forzosamente',
  controls_taken_steal: 'tomó los controles de {name}',
  controls_released: '{name} liberó los controles',
  controls_released_force: 'liberó los controles forzosamente',
  controls_released_steal: 'liberó los controles de {name}',
  controls_given: 'dio los controles a {name}',
  controls_has: '{name} tiene los controles',
  controls_has_alt: 'Pero le hice saber que los querías',
  controls_requesting: '{name} está solicitando los controles',
  resolution: 'cambió la resolución a {width}x{height}@{rate}',
  banned: 'baneó a {name}',
  kicked: 'expulsó a {name}',
  muted: 'silenció a {name}',
  unmuted: 'activó el sonido de {name}',
}

export const files = {
  downloads: 'Descargas',
  uploads: 'Subidas',
  upload_here: 'Haz clic o arrastra archivos aquí para subir',
}
