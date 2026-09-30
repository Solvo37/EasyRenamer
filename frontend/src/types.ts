export type ThemeMode = 'system' | 'dark' | 'light'

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
}

export interface BackendApp {
  Bootstrap(): Promise<BootstrapData>
  SetLanguage(lang: string): Promise<Record<string, string>>
  NewMethod(kind: string): Promise<RenameMethod>
  PickFiles(): Promise<string[]>
  PickFolders(): Promise<string[]>
  Preview(
    sources: string[],
    recursive: boolean,
    category: string,
    customExtensions: string,
    methods: RenameMethod[]
  ): Promise<PreviewResult>
  CancelPreview(): Promise<void>
  Execute(selectedPaths: string[]): Promise<OperationResult>
  Undo(): Promise<OperationResult>
  Reveal(path: string): Promise<void>
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
