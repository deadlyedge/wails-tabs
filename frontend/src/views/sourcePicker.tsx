type SourcePickerProps = {
  sources: string[]
}

export function SourcePicker({ sources }: SourcePickerProps) {
  return (
    <div className='flex flex-col rounded border border-slate-700/40 p-3 gap-2'>
      <div className='text-sm font-semibold text-slate-300'>Source Folders</div>
      {sources.length === 0 ? (
        <span className='text-xs text-slate-500'>No sources configured</span>
      ) : (
        <ul className='text-xs text-slate-200 space-y-1'>
          {sources.map((source) => (
            <li key={source} className='truncate'>
              {source}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
