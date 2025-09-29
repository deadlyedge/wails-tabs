type RulesProps = {
  extensions: string[]
}

export function Rules({ extensions }: RulesProps) {
  return (
    <div className='flex flex-col rounded border border-slate-700/40 p-3 gap-2'>
      <div className='text-sm font-semibold text-slate-300'>Extensions</div>
      <div className='flex flex-wrap gap-2'>
        {extensions.length === 0 ? (
          <span className='text-xs text-slate-500'>Defaults applied</span>
        ) : (
          extensions.map((ext) => (
            <span
              key={ext}
              className='px-2 py-1 rounded bg-slate-700/60 text-xs uppercase tracking-wide text-slate-100'
            >
              {ext.replace(/^\./, "")}
            </span>
          ))
        )}
      </div>
    </div>
  )
}
