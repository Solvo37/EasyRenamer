export type ThemeMode = 'system' | 'dark' | 'light'
export type SortMode = 'name' | 'created' | 'modified' | 'size' | 'extension' | 'path' | 'added' | 'manual'
export type CollisionPolicy = 'skip' | 'auto-number' | 'overwrite' | 'stop'

export interface RenameMethod {
  type: string
  disabled?: boolean
  template?: string
  list_text?: string
  list_include_extension?: boolean
  list_replace_text?: string
  list_replace_regex?: boolean
  list_replace_case_sensitive?: boolean
  find?: string
  replace_with?: string
  use_regex?: boolean
  prefix?: string
  suffix?: string
  case_mode?: 'lower' | 'upper' | 'title'
  remove_start?: number
  remove_count?: number
  remove_pattern?: string
  remove_pattern_regex?: boolean
  renumber_start?: number
  renumber_step?: number
  renumber_padding?: number
  renumber_per_dir?: boolean
  renumber_position?: 'prefix' | 'suffix'
  renumber_separator?: string
  trim_normalize_spaces?: boolean
  timestamp_source?: 'modified' | 'batch'
  timestamp_format?: string
  timestamp_position?: 'prefix' | 'suffix'
  timestamp_separator?: string
  move_start?: number
  move_count?: number
  move_to?: number
  swap_separator?: string
  swap_occurrence?: number
  script_expression?: string
}

export interface PreviewItem {
  sourcePath: string
  oldName: string
  newName: string
  path: string
  status: string
  error?: string
  size: number
  width: number
  height: number
  globalIndex: number
  type: string
}

export interface PreviewResult {
  items: PreviewItem[]
}

export interface FileDetails {
  size: number
  width: number
  height: number
  type: string
}

export interface BootstrapData {
  version: string
  language: string
  translations: Record<string, string>
}

export interface OperationPair {
  from: string
  to: string
}

export interface OperationResult {
  count: number
  pairs?: OperationPair[]
  cancelled?: boolean
}

export interface ExecuteProgress {
  phase: 'staging' | 'renaming'
  completed: number
  total: number
  current: string
}

export interface PathClassification {
  files: string[]
  folders: string[]
}

export interface HistoryEntry {
  id: string
  createdAt: string
  count: number
  folder: string
  undone: boolean
}

export interface BackendApp {
  Bootstrap(): Promise<BootstrapData>
  SetLanguage(lang: string): Promise<Record<string, string>>
  NewMethod(kind: string): Promise<RenameMethod>
  PickFiles(): Promise<string[]>
  PickFolders(): Promise<string[]>
  ClassifyPaths(paths: string[]): Promise<PathClassification>
  Preview(
    sources: string[],
    excludedPaths: string[],
    recursive: boolean,
    category: string,
    customExtensions: string,
    methods: RenameMethod[],
    sortBy: SortMode,
    sortDescending: boolean,
    sortPerFolder: boolean,
    collisionPolicy: CollisionPolicy
  ): Promise<PreviewResult>
  CancelPreview(): Promise<void>
  Execute(selectedPaths: string[]): Promise<OperationResult>
  CancelExecute(): Promise<void>
  Undo(): Promise<OperationResult>
  History(): Promise<HistoryEntry[]>
  UndoHistory(id: string): Promise<OperationResult>
  Reveal(path: string): Promise<void>
  Open(path: string): Promise<void>
  OpenFolder(path: string): Promise<void>
  FileDetails(path: string): Promise<FileDetails>
  Thumbnail(path: string): Promise<string>
  SaveMethodSet(methods: RenameMethod[]): Promise<void>
  LoadMethodSet(): Promise<RenameMethod[]>
}

export interface WailsRuntime {
  WindowMinimise(): void
  WindowToggleMaximise(): void
  Quit(): void
  OnFileDrop(callback: (x: number, y: number, paths: string[]) => void, useDropTarget: boolean): void
  OnFileDropOff(): void
  EventsOn?(eventName: string, callback: (data: any) => void): () => void
  EventsOff?(eventName: string): void
}

declare global {
  interface Window {
    go?: {
      main?: {
        App?: BackendApp
      }
    }
    runtime?: WailsRuntime
  }
}
