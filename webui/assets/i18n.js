/* NET PORTAL — interface strings. {0}, {1}… are placeholders. */
'use strict';

const I18N = {
  ru: {
    'nav.tunnels': 'Туннели', 'nav.log': 'Журнал', 'nav.settings': 'Настройки', 'nav.about': 'О программе',
    'update.available': 'Доступно обновление', 'update.short': 'Новая версия готова к установке', 'update.button': 'Обновить',
    'demo': 'Демо-режим: данные ненастоящие, службы не затрагиваются',

    'list.title': 'Туннели', 'list.add': 'Добавить', 'list.search': 'Поиск', 'list.nothing': 'Ничего не найдено',
    'list.exportAll': 'Экспорт всех в .zip', 'list.multi': 'Выбрать несколько', 'list.removeSel': 'Удалить выбранный туннель',
    'list.selected': 'Выбрано: {0}', 'list.export': 'Экспорт', 'list.remove': 'Удалить', 'list.cancel': 'Готово',
    'list.toggle': 'Подключить {0}', 'list.select': 'Выбрать {0}', 'list.others': 'ещё {0}',

    'type.full': 'Полный маршрут', 'type.partial': 'Частичный',
    'st.connected': 'Активен', 'st.connecting': 'Подключение…', 'st.disconnecting': 'Отключение…', 'st.error': 'Ошибка подключения', 'st.idle': 'Неактивен', 'st.unknown': 'Неизвестно',
    'row.connected': 'Активен · {0}', 'row.connectedNoTime': 'Активен',

    'hero.sub.idle': 'Нажмите кнопку, чтобы подключиться',
    'hero.sub.connecting': 'Устанавливаем соединение с {0}',
    'hero.sub.connectingNoEp': 'Устанавливаем соединение',
    'hero.sub.connected.full.ks': 'Весь трафик идёт через туннель · kill-switch включён',
    'hero.sub.connected.full': 'Весь трафик идёт через туннель',
    'hero.sub.connected.partial': 'Через туннель идёт только трафик к выбранным сетям',
    'hero.sub.disconnecting': 'Закрываем туннель и восстанавливаем маршруты',
    'hero.sub.error': 'Не удалось подключиться',
    'hero.action.idle': 'Подключить', 'hero.action.connecting': 'Отменить подключение', 'hero.action.connected': 'Отключить',
    'hero.action.disconnecting': 'Отключение', 'hero.action.error': 'Повторить подключение',
    'hero.endpoint': 'Endpoint', 'hero.addr': 'Адрес в туннеле', 'hero.extip': 'Внешний IP',
    'hero.extip.off': 'проверка отключена', 'hero.extip.loading': 'определяем…', 'hero.extip.failed': 'не удалось определить',

    'err.title': 'Не удалось установить соединение',
    'err.hint': 'Проверьте endpoint и параметры обфускации — они должны совпадать с сервером.',
    'err.retry': 'Повторить', 'err.openLog': 'Открыть журнал',

    'stats.down': 'Загрузка', 'stats.up': 'Отдача', 'stats.window': 'последние 60 с', 'stats.mbps': 'Мбит/с',
    'stats.received': 'Получено', 'stats.sent': 'Отправлено', 'stats.handshake': 'Последний handshake',
    'stats.wait': 'Ожидаем первый handshake…', 'stats.collecting': 'Собираем статистику…', 'stats.after': 'Статистика появится после подключения',
    'stats.graph': 'График скорости за последние 60 секунд',

    'iface.title': 'Интерфейс', 'iface.publicKey': 'Публичный ключ', 'iface.port': 'Порт прослушивания', 'iface.mtu': 'MTU',
    'iface.addresses': 'Адреса', 'iface.dns': 'DNS-серверы', 'iface.table': 'Таблица маршрутизации', 'iface.scripts': 'Скрипты',
    'iface.auto': 'auto', 'iface.off': 'off',

    'obf.title': 'Обфускация AmneziaWG', 'obf.badge': 'Для продвинутых', 'obf.count': 'Параметров: {0}', 'obf.none': 'не настроена',
    'obf.hint': 'Параметры маскируют трафик под случайный UDP. Значения должны совпадать с настройками сервера — иначе handshake не пройдёт.',
    'obf.set': 'задан',
    'obf.Jc': 'Мусорные пакеты (Jc)', 'obf.Jmin': 'Мин. размер мусора (Jmin)', 'obf.Jmax': 'Макс. размер мусора (Jmax)',
    'obf.S1': 'Паддинг init (S1)', 'obf.S2': 'Паддинг response (S2)', 'obf.S3': 'Паддинг cookie (S3)', 'obf.S4': 'Паддинг transport (S4)',
    'obf.H1': 'Заголовок init (H1)', 'obf.H2': 'Заголовок response (H2)', 'obf.H3': 'Заголовок cookie (H3)', 'obf.H4': 'Заголовок transport (H4)',
    'obf.HeaderProtectionKey': 'Ключ защиты заголовков', 'obf.ContentPaddingAddition': 'Доп. паддинг контента',
    'obf.RekeyAfterTime': 'Rekey after time', 'obf.RekeyTimeout': 'Rekey timeout', 'obf.RejectAfterTime': 'Reject after time',
    'obf.KeepaliveTimeout': 'Keepalive timeout', 'obf.MaxHandshakeAttempts': 'Макс. попыток handshake',
    'obf.RandomTrailers': 'Случайные трейлеры', 'obf.DisableCookies': 'Отключение cookies',

    'peer.title': 'Пир {0}', 'peer.publicKey': 'Публичный ключ', 'peer.psk': 'Preshared key', 'peer.endpoint': 'Endpoint',
    'peer.allowed': 'Allowed IPs', 'peer.keepalive': 'Persistent keepalive', 'peer.traffic': 'Трафик',
    'peer.show': 'Показать', 'peer.hide': 'Скрыть', 'peer.none': 'не задан', 'peer.hidden': 'задан (виден администраторам)',
    'peer.every': 'каждые {0} с', 'peer.off': 'выкл.', 'peer.noData': 'нет данных', 'peer.never': 'ещё не было',

    'det.edit': 'Редактировать', 'det.export': 'Экспорт туннеля', 'det.delete': 'Удалить туннель',
    'det.readonly': 'Изменять туннели могут только администраторы этого компьютера.',

    'copy': 'Копировать', 'copied': 'Скопировано', 'copiedText': '{0} скопирован в буфер обмена',

    'empty.title': 'Пока нет ни одного туннеля',
    'empty.text': 'Перетащите файл .conf или архив .zip в это окно — или выберите способ ниже.',
    'empty.import': 'Импорт из файла…', 'empty.paste': 'Вставить из буфера', 'empty.qr': 'Сканировать QR', 'empty.new': 'Создать пустой',
    'empty.support': 'Поддерживаются конфигурации WireGuard и AmneziaWG',
    'empty.readonly': 'Туннели может добавить администратор этого компьютера.',

    'add.file': 'Импорт из файла…', 'add.file.hint': '.conf или .zip с несколькими туннелями',
    'add.paste': 'Вставить из буфера', 'add.paste.hint': 'Текст конфигурации [Interface]…',
    'add.qr': 'Сканировать QR-код', 'add.qr.hint': 'С экрана или из изображения',
    'add.qr.screen': 'С экрана', 'add.qr.file': 'Из изображения…',
    'add.new': 'Создать пустой туннель', 'add.new.hint': 'Сгенерировать новую пару ключей',
    'dlg.import': 'Импорт туннелей из файла', 'dlg.exportAll': 'Экспорт туннелей в архив', 'dlg.exportOne': 'Экспорт туннеля',
    'dlg.saveLog': 'Сохранить журнал', 'dlg.qr': 'Изображение с QR-кодом',

    'del.title.one': 'Удалить туннель «{0}»?', 'del.title.many': 'Удалить {0}?',
    'del.text': 'Это действие нельзя отменить. Конфигурация и ключи будут удалены с этого компьютера.',
    'del.confirm': 'Удалить', 'cancel': 'Отмена',

    'sw.title': 'Переключиться на «{0}»?',
    'sw.text.one': 'Сейчас подключён туннель {0} — он будет отключён. Пока подключается «{1}», соединение на несколько секунд прервётся: загрузки и звонки могут оборваться.',
    'sw.text.many': 'Сейчас подключены туннели {0} — они будут отключены. Пока подключается «{1}», соединение на несколько секунд прервётся: загрузки и звонки могут оборваться.',
    'sw.dontAsk': 'Больше не спрашивать', 'sw.confirm': 'Переключиться',
    'set.switchWarn': 'Предупреждать при переключении', 'set.switchWarn.hint': 'Спрашивать перед подключением туннеля, который отключит текущий',

    'drop.title': 'Отпустите, чтобы импортировать', 'drop.sub': '.conf и .zip',

    'log.title': 'Журнал', 'log.sub': 'Записей: {0}', 'log.copy': 'Копировать', 'log.save': 'Сохранить в файл…',
    'log.filter': 'Фильтр по тексту', 'log.auto': 'Автопрокрутка',
    'log.all': 'Все', 'log.info': 'Инфо', 'log.warn': 'Предупреждения', 'log.error': 'Ошибки',
    'log.col.time': 'Время', 'log.col.level': 'Уровень', 'log.col.msg': 'Сообщение',
    'log.lv.info': 'ИНФО', 'log.lv.warn': 'ВНИМ.', 'log.lv.error': 'ОШИБКА', 'log.empty': 'Нет записей',

    'set.title': 'Настройки', 'set.look': 'Оформление', 'set.theme': 'Тема', 'set.theme.hint': 'Заголовок окна подстраивается под тему',
    'set.theme.dark': 'Тёмная', 'set.theme.light': 'Светлая', 'set.theme.system': 'Как в системе',
    'set.lang': 'Язык интерфейса', 'set.lang.hint': 'Применяется сразу',
    'set.startup': 'Запуск и подключение',
    'set.autostart': 'Автозапуск с Windows', 'set.autostart.hint': 'Запускать NET PORTAL при входе в систему',
    'set.autostart.admin': 'Изменить может только администратор',
    'set.minimized': 'Запускать свёрнутым в трей', 'set.minimized.hint': 'Окно не появится при автозапуске',
    'set.autoconnect': 'Автоподключение', 'set.autoconnect.hint': 'Подключаться к выбранному туннелю при запуске',
    'set.notify': 'Уведомления', 'set.notify.hint': 'Сообщать о подключении, отключении и ошибках',
    'set.privacy': 'Приватность',
    'set.extip': 'Показывать внешний IP', 'set.extip.hint': 'После подключения запрашивает адрес и страну у 2ip (api.2ip.io)',
    'set.noTunnel': 'не выбран',

    'about.tagline': 'VPN-клиент на базе AmneziaWG с обфускацией трафика',
    'about.app': 'Приложение', 'about.wintun': 'Wintun', 'about.go': 'Go', 'about.os': 'ОС', 'about.arch': 'Архитектура',
    'about.upd.title': 'Доступна новая версия', 'about.upd.sub': 'Рекомендуется обновиться как можно скорее',
    'about.upd.now': 'Обновить сейчас', 'about.upd.installing': 'Устанавливаем обновление',
    'about.upd.restart': 'Туннели переподключатся автоматически после перезапуска',
    'about.upd.progress': '{0} · {1} из {2}', 'about.upd.done': 'Готово', 'about.upd.error': 'Не удалось обновить: {0}',
    'about.unofficial': 'Автообновление не настроено в этой сборке', 'about.upd.restarting': 'Обновление установлено, перезапускаем…',

    'ed.crumb': 'Туннели / {0}', 'ed.crumbNew': 'Туннели / новый', 'ed.title.edit': 'Редактирование туннеля', 'ed.title.new': 'Новый туннель',
    'ed.unsaved': 'Есть несохранённые изменения', 'ed.name': 'Имя', 'ed.name.hint': 'Латиница, цифры, «-», «_», «=», «+», «.», до 32 символов',
    'ed.pub': 'Публичный ключ', 'ed.pub.hint': 'Вычисляется из PrivateKey, только для чтения', 'ed.pub.unknown': '(неизвестно)',
    'ed.config': 'Конфигурация', 'ed.meta': 'INI · UTF-8', 'ed.pos': 'Стр. {0}, стб. {1}',
    'ed.ks': 'Блокировать трафик вне туннеля (kill-switch)',
    'ed.ks.hint': 'Если туннель упадёт, Windows не отправит ни одного пакета в обход VPN. Доступно при одном пире с полным маршрутом.',
    'ed.keys': 'Ctrl+S — сохранить · Esc — отмена', 'ed.save': 'Сохранить', 'ed.fix': 'Исправить',
    'ed.line': 'Строка {0}:', 'ed.unknownKey': 'неизвестный ключ «{0}».', 'ed.didYouMean': 'Возможно, вы имели в виду «{0}»?',
    'ed.badValue': 'неверное значение «{0}».', 'ed.badSection': 'неизвестная секция «{0}».', 'ed.warning': 'подозрительное значение «{0}» — проверьте Jc, Jmin, Jmax и MTU.',
    'ed.leave': 'Отменить изменения?', 'ed.leave.text': 'Несохранённые изменения будут потеряны.', 'ed.leave.ok': 'Не сохранять',

    'code.nameRequired': 'Укажите имя туннеля', 'code.nameInvalid': 'Недопустимое имя туннеля',
    'code.nameExists': 'Туннель с таким именем уже существует', 'code.parse': 'Ошибка в конфигурации: {0}',
    'code.notAdmin': 'Требуются права администратора', 'code.noConfigs': 'Файлы конфигурации не найдены',
    'code.notAConfig': 'В буфере нет конфигурации — ожидается текст с секцией [Interface]',
    'code.vpnLinkUnsupported': 'Ссылки vpn:// пока не поддерживаются — экспортируйте конфигурацию AmneziaWG в формате .conf',
    'code.noQR': 'QR-код не найден', 'code.clipboard': 'Не удалось прочитать буфер обмена',
    'code.autostartFailed': 'Не удалось изменить автозапуск: {0}', 'code.generic': 'Ошибка: {0}',

    'toast.connected': 'Туннель {0} подключён', 'toast.connected.ip': 'Внешний IP {0}',
    'toast.disconnected': 'Туннель {0} отключён', 'toast.error': 'Ошибка туннеля {0}',
    'toast.imported': 'Импортировано: {0}', 'toast.importedSome': 'Импортировано {0} из {1}',
    'toast.importFailed': 'Не удалось импортировать', 'toast.saved': 'Туннель {0} сохранён', 'toast.deleted': 'Удалено: {0}',
    'toast.exported': 'Экспорт завершён', 'toast.logSaved': 'Журнал сохранён',
    'toast.qrFound': 'QR-код распознан', 'toast.qrFound.text': 'Проверьте конфигурацию и сохраните туннель',
    'toast.close': 'Закрыть уведомление',

    'tray.disconnected': 'Отключено', 'tray.connecting': 'Подключение…', 'tray.connected': 'Подключено', 'tray.error': 'Ошибка',
    'tray.last': 'Последний: {0}', 'tray.none': 'Нет туннелей', 'tray.tunnels': 'Туннели',
    'tray.down': 'Загрузка', 'tray.up': 'Отдача', 'tray.time': 'Время', 'tray.open': 'Открыть NET PORTAL', 'tray.exit': 'Выход',
    'tray.settings': 'Настройки', 'tray.active': 'активен', 'tray.connectingShort': 'подключение…', 'tray.toggle': 'Подключение к {0}',

    'time.now': 'только что', 'time.sAgo': '{0} с назад', 'time.mAgo': '{0} мин {1} с назад', 'time.hAgo': '{0} ч {1} мин назад', 'time.dAgo': '{0} дн. назад',
    'unit.B': 'Б', 'unit.KiB': 'КиБ', 'unit.MiB': 'МиБ', 'unit.GiB': 'ГиБ', 'unit.TiB': 'ТиБ',
    'tunnels.one': '{0} туннель', 'tunnels.few': '{0} туннеля', 'tunnels.many': '{0} туннелей',
  },

  en: {
    'nav.tunnels': 'Tunnels', 'nav.log': 'Log', 'nav.settings': 'Settings', 'nav.about': 'About',
    'update.available': 'Update available', 'update.short': 'A new version is ready to install', 'update.button': 'Update',
    'demo': 'Demo mode: sample data, no services are touched',

    'list.title': 'Tunnels', 'list.add': 'Add', 'list.search': 'Search', 'list.nothing': 'Nothing found',
    'list.exportAll': 'Export all to .zip', 'list.multi': 'Select several', 'list.removeSel': 'Delete selected tunnel',
    'list.selected': 'Selected: {0}', 'list.export': 'Export', 'list.remove': 'Delete', 'list.cancel': 'Done',
    'list.toggle': 'Connect {0}', 'list.select': 'Select {0}', 'list.others': '{0} more',

    'type.full': 'Full route', 'type.partial': 'Split',
    'st.connected': 'Active', 'st.connecting': 'Connecting…', 'st.disconnecting': 'Disconnecting…', 'st.error': 'Connection failed', 'st.idle': 'Inactive', 'st.unknown': 'Unknown',
    'row.connected': 'Active · {0}', 'row.connectedNoTime': 'Active',

    'hero.sub.idle': 'Press the button to connect',
    'hero.sub.connecting': 'Establishing a connection to {0}',
    'hero.sub.connectingNoEp': 'Establishing a connection',
    'hero.sub.connected.full.ks': 'All traffic goes through the tunnel · kill-switch on',
    'hero.sub.connected.full': 'All traffic goes through the tunnel',
    'hero.sub.connected.partial': 'Only traffic to the selected networks goes through the tunnel',
    'hero.sub.disconnecting': 'Closing the tunnel and restoring routes',
    'hero.sub.error': 'Could not connect',
    'hero.action.idle': 'Connect', 'hero.action.connecting': 'Cancel connection', 'hero.action.connected': 'Disconnect',
    'hero.action.disconnecting': 'Disconnecting', 'hero.action.error': 'Retry connection',
    'hero.endpoint': 'Endpoint', 'hero.addr': 'Tunnel address', 'hero.extip': 'External IP',
    'hero.extip.off': 'check disabled', 'hero.extip.loading': 'detecting…', 'hero.extip.failed': 'unavailable',

    'err.title': 'Could not establish a connection',
    'err.hint': 'Check the endpoint and the obfuscation parameters — they must match the server.',
    'err.retry': 'Retry', 'err.openLog': 'Open log',

    'stats.down': 'Download', 'stats.up': 'Upload', 'stats.window': 'last 60 s', 'stats.mbps': 'Mbit/s',
    'stats.received': 'Received', 'stats.sent': 'Sent', 'stats.handshake': 'Latest handshake',
    'stats.wait': 'Waiting for the first handshake…', 'stats.collecting': 'Collecting statistics…', 'stats.after': 'Statistics appear once connected',
    'stats.graph': 'Speed over the last 60 seconds',

    'iface.title': 'Interface', 'iface.publicKey': 'Public key', 'iface.port': 'Listen port', 'iface.mtu': 'MTU',
    'iface.addresses': 'Addresses', 'iface.dns': 'DNS servers', 'iface.table': 'Routing table', 'iface.scripts': 'Scripts',
    'iface.auto': 'auto', 'iface.off': 'off',

    'obf.title': 'AmneziaWG obfuscation', 'obf.badge': 'Advanced', 'obf.count': '{0} parameters', 'obf.none': 'not configured',
    'obf.hint': 'These parameters disguise the traffic as random UDP. They must match the server, or the handshake will fail.',
    'obf.set': 'set',
    'obf.Jc': 'Junk packets (Jc)', 'obf.Jmin': 'Junk min size (Jmin)', 'obf.Jmax': 'Junk max size (Jmax)',
    'obf.S1': 'Init padding (S1)', 'obf.S2': 'Response padding (S2)', 'obf.S3': 'Cookie padding (S3)', 'obf.S4': 'Transport padding (S4)',
    'obf.H1': 'Init header (H1)', 'obf.H2': 'Response header (H2)', 'obf.H3': 'Cookie header (H3)', 'obf.H4': 'Transport header (H4)',
    'obf.HeaderProtectionKey': 'Header protection key', 'obf.ContentPaddingAddition': 'Content padding addition',
    'obf.RekeyAfterTime': 'Rekey after time', 'obf.RekeyTimeout': 'Rekey timeout', 'obf.RejectAfterTime': 'Reject after time',
    'obf.KeepaliveTimeout': 'Keepalive timeout', 'obf.MaxHandshakeAttempts': 'Max handshake attempts',
    'obf.RandomTrailers': 'Random trailers', 'obf.DisableCookies': 'Disable cookies',

    'peer.title': 'Peer {0}', 'peer.publicKey': 'Public key', 'peer.psk': 'Preshared key', 'peer.endpoint': 'Endpoint',
    'peer.allowed': 'Allowed IPs', 'peer.keepalive': 'Persistent keepalive', 'peer.traffic': 'Traffic',
    'peer.show': 'Show', 'peer.hide': 'Hide', 'peer.none': 'not set', 'peer.hidden': 'set (visible to administrators)',
    'peer.every': 'every {0} s', 'peer.off': 'off', 'peer.noData': 'no data', 'peer.never': 'never',

    'det.edit': 'Edit', 'det.export': 'Export tunnel', 'det.delete': 'Delete tunnel',
    'det.readonly': 'Only administrators of this computer can change tunnels.',

    'copy': 'Copy', 'copied': 'Copied', 'copiedText': '{0} copied to the clipboard',

    'empty.title': 'No tunnels yet',
    'empty.text': 'Drop a .conf file or a .zip archive into this window — or pick an option below.',
    'empty.import': 'Import from file…', 'empty.paste': 'Paste from clipboard', 'empty.qr': 'Scan QR', 'empty.new': 'Create empty',
    'empty.support': 'WireGuard and AmneziaWG configurations are supported',
    'empty.readonly': 'An administrator of this computer can add tunnels.',

    'add.file': 'Import from file…', 'add.file.hint': '.conf or a .zip with several tunnels',
    'add.paste': 'Paste from clipboard', 'add.paste.hint': 'Configuration text with [Interface]…',
    'add.qr': 'Scan QR code', 'add.qr.hint': 'From the screen or an image',
    'add.qr.screen': 'From screen', 'add.qr.file': 'From image…',
    'add.new': 'Create empty tunnel', 'add.new.hint': 'Generate a new key pair',
    'dlg.import': 'Import tunnels from file', 'dlg.exportAll': 'Export tunnels to archive', 'dlg.exportOne': 'Export tunnel',
    'dlg.saveLog': 'Save log', 'dlg.qr': 'Image with a QR code',

    'del.title.one': 'Delete tunnel “{0}”?', 'del.title.many': 'Delete {0}?',
    'del.text': 'This cannot be undone. The configuration and keys will be removed from this computer.',
    'del.confirm': 'Delete', 'cancel': 'Cancel',

    'sw.title': 'Switch to “{0}”?',
    'sw.text.one': 'Tunnel {0} is connected and will be disconnected. While “{1}” connects, the connection drops for a few seconds: downloads and calls may be interrupted.',
    'sw.text.many': 'Tunnels {0} are connected and will be disconnected. While “{1}” connects, the connection drops for a few seconds: downloads and calls may be interrupted.',
    'sw.dontAsk': 'Don’t ask again', 'sw.confirm': 'Switch',
    'set.switchWarn': 'Warn when switching', 'set.switchWarn.hint': 'Ask before connecting a tunnel that disconnects the current one',

    'drop.title': 'Release to import', 'drop.sub': '.conf and .zip',

    'log.title': 'Log', 'log.sub': '{0} entries', 'log.copy': 'Copy', 'log.save': 'Save to file…',
    'log.filter': 'Filter text', 'log.auto': 'Auto-scroll',
    'log.all': 'All', 'log.info': 'Info', 'log.warn': 'Warnings', 'log.error': 'Errors',
    'log.col.time': 'Time', 'log.col.level': 'Level', 'log.col.msg': 'Message',
    'log.lv.info': 'INFO', 'log.lv.warn': 'WARN', 'log.lv.error': 'ERROR', 'log.empty': 'No entries',

    'set.title': 'Settings', 'set.look': 'Appearance', 'set.theme': 'Theme', 'set.theme.hint': 'The window title bar follows the theme',
    'set.theme.dark': 'Dark', 'set.theme.light': 'Light', 'set.theme.system': 'System',
    'set.lang': 'Language', 'set.lang.hint': 'Applies immediately',
    'set.startup': 'Startup and connection',
    'set.autostart': 'Start with Windows', 'set.autostart.hint': 'Launch NET PORTAL when you sign in',
    'set.autostart.admin': 'Only an administrator can change this',
    'set.minimized': 'Start minimized to tray', 'set.minimized.hint': 'The window does not appear on autostart',
    'set.autoconnect': 'Auto-connect', 'set.autoconnect.hint': 'Connect to the selected tunnel on start',
    'set.notify': 'Notifications', 'set.notify.hint': 'Tell me about connections, disconnections and errors',
    'set.privacy': 'Privacy',
    'set.extip': 'Show external IP', 'set.extip.hint': 'After connecting, asks 2ip (api.2ip.io) for the address and country',
    'set.noTunnel': 'none',

    'about.tagline': 'AmneziaWG-based VPN client with traffic obfuscation',
    'about.app': 'App', 'about.wintun': 'Wintun', 'about.go': 'Go', 'about.os': 'OS', 'about.arch': 'Architecture',
    'about.upd.title': 'A new version is available', 'about.upd.sub': 'Updating as soon as possible is recommended',
    'about.upd.now': 'Update now', 'about.upd.installing': 'Installing the update',
    'about.upd.restart': 'Tunnels reconnect automatically after the restart',
    'about.upd.progress': '{0} · {1} of {2}', 'about.upd.done': 'Done', 'about.upd.error': 'Update failed: {0}',
    'about.unofficial': 'Automatic updates are not configured in this build', 'about.upd.restarting': 'Update installed, restarting…',

    'ed.crumb': 'Tunnels / {0}', 'ed.crumbNew': 'Tunnels / new', 'ed.title.edit': 'Edit tunnel', 'ed.title.new': 'New tunnel',
    'ed.unsaved': 'Unsaved changes', 'ed.name': 'Name', 'ed.name.hint': 'Letters, digits, “-”, “_”, “=”, “+”, “.”, up to 32 characters',
    'ed.pub': 'Public key', 'ed.pub.hint': 'Derived from PrivateKey, read-only', 'ed.pub.unknown': '(unknown)',
    'ed.config': 'Configuration', 'ed.meta': 'INI · UTF-8', 'ed.pos': 'Ln {0}, Col {1}',
    'ed.ks': 'Block untunneled traffic (kill-switch)',
    'ed.ks.hint': 'If the tunnel drops, Windows will not send a single packet outside the VPN. Available with a single full-route peer.',
    'ed.keys': 'Ctrl+S — save · Esc — cancel', 'ed.save': 'Save', 'ed.fix': 'Fix',
    'ed.line': 'Line {0}:', 'ed.unknownKey': 'unknown key “{0}”.', 'ed.didYouMean': 'Did you mean “{0}”?',
    'ed.badValue': 'invalid value “{0}”.', 'ed.badSection': 'unknown section “{0}”.', 'ed.warning': 'suspicious value “{0}” — check Jc, Jmin, Jmax and MTU.',
    'ed.leave': 'Discard changes?', 'ed.leave.text': 'Unsaved changes will be lost.', 'ed.leave.ok': 'Discard',

    'code.nameRequired': 'Enter a tunnel name', 'code.nameInvalid': 'Invalid tunnel name',
    'code.nameExists': 'A tunnel with this name already exists', 'code.parse': 'Configuration error: {0}',
    'code.notAdmin': 'Administrator rights are required', 'code.noConfigs': 'No configuration files were found',
    'code.notAConfig': 'The clipboard has no configuration — expected text with an [Interface] section',
    'code.vpnLinkUnsupported': 'vpn:// links are not supported yet — export the AmneziaWG configuration as .conf',
    'code.noQR': 'No QR code found', 'code.clipboard': 'Could not read the clipboard',
    'code.autostartFailed': 'Could not change autostart: {0}', 'code.generic': 'Error: {0}',

    'toast.connected': 'Tunnel {0} connected', 'toast.connected.ip': 'External IP {0}',
    'toast.disconnected': 'Tunnel {0} disconnected', 'toast.error': 'Tunnel {0} error',
    'toast.imported': 'Imported: {0}', 'toast.importedSome': 'Imported {0} of {1}',
    'toast.importFailed': 'Import failed', 'toast.saved': 'Tunnel {0} saved', 'toast.deleted': 'Deleted: {0}',
    'toast.exported': 'Export complete', 'toast.logSaved': 'Log saved',
    'toast.qrFound': 'QR code recognized', 'toast.qrFound.text': 'Review the configuration and save the tunnel',
    'toast.close': 'Dismiss notification',

    'tray.disconnected': 'Disconnected', 'tray.connecting': 'Connecting…', 'tray.connected': 'Connected', 'tray.error': 'Error',
    'tray.last': 'Last: {0}', 'tray.none': 'No tunnels', 'tray.tunnels': 'Tunnels',
    'tray.down': 'Download', 'tray.up': 'Upload', 'tray.time': 'Time', 'tray.open': 'Open NET PORTAL', 'tray.exit': 'Exit',
    'tray.settings': 'Settings', 'tray.active': 'active', 'tray.connectingShort': 'connecting…', 'tray.toggle': 'Connection to {0}',

    'time.now': 'just now', 'time.sAgo': '{0} s ago', 'time.mAgo': '{0} min {1} s ago', 'time.hAgo': '{0} h {1} min ago', 'time.dAgo': '{0} d ago',
    'unit.B': 'B', 'unit.KiB': 'KiB', 'unit.MiB': 'MiB', 'unit.GiB': 'GiB', 'unit.TiB': 'TiB',
    'tunnels.one': '{0} tunnel', 'tunnels.few': '{0} tunnels', 'tunnels.many': '{0} tunnels',
  },
};

let LANG = 'ru';

function t(key, ...args) {
  const s = (I18N[LANG] && I18N[LANG][key]) || I18N.ru[key] || key;
  return args.length ? s.replace(/\{(\d+)\}/g, (m, i) => (args[i] === undefined ? m : String(args[i]))) : s;
}

// Russian plural forms; English uses one/many.
function plural(n, base) {
  let form = 'many';
  if (LANG === 'ru') {
    const m10 = n % 10, m100 = n % 100;
    if (m10 === 1 && m100 !== 11) form = 'one';
    else if (m10 >= 2 && m10 <= 4 && (m100 < 12 || m100 > 14)) form = 'few';
  } else if (n === 1) form = 'one';
  return t(base + '.' + form, n);
}
