type TargetPickerProps = {
  targetPath: string
  pattern: string
}

export function TargetPicker({ targetPath, pattern }: TargetPickerProps) {
  return (
    <div className='flex flex-col rounded border border-slate-700/40 p-3 gap-2'>
      <div className='text-sm font-semibold text-slate-300'>Target Layout</div>
      <div className='text-xs text-slate-200 truncate'>Base: {targetPath || "(not set)"}</div>
      <div className='text-xs text-slate-400'>Pattern: {pattern}</div>
    </div>
  )
}
