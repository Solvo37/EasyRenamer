export interface MethodDefinition {
  type: string
  titleKey: string
  descriptionKey: string
  group: 'common' | 'advanced'
  icon: 'pen' | 'replace' | 'hash' | 'case' | 'remove' | 'move' | 'list' | 'script' | 'swap' | 'trim' | 'time' | 'add'
}

export const methodCatalog: MethodDefinition[] = [
  { type: 'template', titleKey: 'method.new_name', descriptionKey: 'method.desc.template', group: 'common', icon: 'pen' },
  { type: 'replace', titleKey: 'method.replace', descriptionKey: 'method.desc.replace', group: 'common', icon: 'replace' },
  { type: 'renumber', titleKey: 'method.renumber', descriptionKey: 'method.desc.renumber', group: 'common', icon: 'hash' },
  { type: 'case', titleKey: 'method.change_case', descriptionKey: 'method.desc.case', group: 'common', icon: 'case' },
  { type: 'remove', titleKey: 'method.remove', descriptionKey: 'method.desc.remove', group: 'common', icon: 'remove' },
  { type: 'prefix_suffix', titleKey: 'method.add_text', descriptionKey: 'method.desc.add_text', group: 'common', icon: 'add' },
  { type: 'move', titleKey: 'method.move', descriptionKey: 'method.desc.move', group: 'advanced', icon: 'move' },
  { type: 'remove_pattern', titleKey: 'method.remove_pattern', descriptionKey: 'method.desc.remove_pattern', group: 'advanced', icon: 'remove' },
  { type: 'list', titleKey: 'method.list', descriptionKey: 'method.desc.list', group: 'advanced', icon: 'list' },
  { type: 'list_replace', titleKey: 'method.list_replace', descriptionKey: 'method.desc.list_replace', group: 'advanced', icon: 'list' },
  { type: 'swap', titleKey: 'method.swap', descriptionKey: 'method.desc.swap', group: 'advanced', icon: 'swap' },
  { type: 'trim', titleKey: 'method.trim', descriptionKey: 'method.desc.trim', group: 'advanced', icon: 'trim' },
  { type: 'timestamp', titleKey: 'method.timestamp', descriptionKey: 'method.desc.timestamp', group: 'advanced', icon: 'time' },
  { type: 'script', titleKey: 'method.script', descriptionKey: 'method.desc.script', group: 'advanced', icon: 'script' }
]

export interface TagEntry {
  token: string
  labelKey: string
}

export interface TagCategory {
  key: string
  items: TagEntry[]
}

export const tagCatalog: TagCategory[] = [
  {
    key: 'tagcat.default',
    items: [
      { token: '<Inc Nr:001:1>', labelKey: 'tag.increment_number' },
      { token: '<Inc NrDir:01:1>', labelKey: 'tag.increment_folder' },
      { token: '<Inc Alpha:A:1>', labelKey: 'tag.increment_letter' },
      { token: '<Name>', labelKey: 'tag.name' },
      { token: '<Ext>', labelKey: 'tag.ext' },
      { token: '<FolderName:1>', labelKey: 'tag.folder_name' },
      { token: '<Num Files:000>', labelKey: 'tag.num_files' },
      { token: '<Num Dirs:000>', labelKey: 'tag.num_dirs' },
      { token: '<Num Items:000>', labelKey: 'tag.num_items' },
      { token: '<Word:1>', labelKey: 'tag.word' },
      { token: '<MediaType>', labelKey: 'tag.media_type' }
    ]
  },
  {
    key: 'tagcat.advanced',
    items: [
      { token: '<Substr:1:5>', labelKey: 'tag.substr' },
      { token: '<RSubstr:1:5>', labelKey: 'tag.rsubstr' },
      { token: '<Switch:A:B>', labelKey: 'tag.switch' },
      { token: '<Rand:1:100>', labelKey: 'tag.random_number' },
      { token: '<Rand Str:8>', labelKey: 'tag.random_string' },
      { token: '<Rand Alpha:8>', labelKey: 'tag.random_letters' },
      { token: '<Inc Hex:1:1>', labelKey: 'tag.increment_hex' },
      { token: '<Inc Roman:1:1>', labelKey: 'tag.increment_roman' },
      { token: '<MetaData:fieldname>', labelKey: 'tag.metadata' }
    ]
  },
  {
    key: 'tagcat.time',
    items: [
      { token: '<Date:yyyy-mm-dd>', labelKey: 'tag.date' },
      { token: '<Time:hh:nn:ss>', labelKey: 'tag.time' },
      { token: '<Date Created:yyyy-mm-dd>', labelKey: 'tag.created_date' },
      { token: '<Date Modified:yyyy-mm-dd>', labelKey: 'tag.modified_date' },
      { token: '<UnixTimestamp>', labelKey: 'tag.unix' }
    ]
  },
  {
    key: 'tagcat.image',
    items: [
      { token: '<Width>', labelKey: 'tag.width' },
      { token: '<Height>', labelKey: 'tag.height' },
      { token: '<Img DateOriginal:yyyy-mm-dd>', labelKey: 'tag.image_date' },
      { token: '<Img DPI>', labelKey: 'tag.dpi' },
      { token: '<Author>', labelKey: 'tag.author' },
      { token: '<Copyright>', labelKey: 'tag.copyright' }
    ]
  },
  {
    key: 'tagcat.audio',
    items: [
      { token: '<Artist>', labelKey: 'tag.artist' },
      { token: '<Album>', labelKey: 'tag.album' },
      { token: '<Title>', labelKey: 'tag.title' },
      { token: '<Genre>', labelKey: 'tag.genre' },
      { token: '<Track:00>', labelKey: 'tag.track' }
    ]
  },
  {
    key: 'tagcat.video',
    items: [
      { token: '<Width>', labelKey: 'tag.width' },
      { token: '<Height>', labelKey: 'tag.height' },
      { token: '<FrameRate>', labelKey: 'tag.framerate' },
      { token: '<Duration>', labelKey: 'tag.duration' },
      { token: '<Video Date:yyyy-mm-dd>', labelKey: 'tag.video_date' }
    ]
  },
  {
    key: 'tagcat.docs',
    items: [
      { token: '<Pages:000>', labelKey: 'tag.pages' },
      { token: '<Creator>', labelKey: 'tag.creator' },
      { token: '<Author>', labelKey: 'tag.author' },
      { token: '<Title>', labelKey: 'tag.title' },
      { token: '<Subject>', labelKey: 'tag.subject' }
    ]
  },
  {
    key: 'tagcat.gps',
    items: [
      { token: '<GPS Lat>', labelKey: 'tag.gps_lat' },
      { token: '<GPS Lng>', labelKey: 'tag.gps_lng' },
      { token: '<GPS Alt>', labelKey: 'tag.gps_alt' }
    ]
  },
  {
    key: 'tagcat.file',
    items: [
      { token: '<Filesize Text>', labelKey: 'tag.filesize_text' },
      { token: '<Filesize B:000>', labelKey: 'tag.filesize_b' },
      { token: '<Filesize Mb:000>', labelKey: 'tag.filesize_mb' },
      { token: '<MD5>', labelKey: 'tag.md5' },
      { token: '<SHA1>', labelKey: 'tag.sha1' },
      { token: '<Exe Product>', labelKey: 'tag.exe_product' },
      { token: '<Exe Version>', labelKey: 'tag.exe_version' }
    ]
  }
]
