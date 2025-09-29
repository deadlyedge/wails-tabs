type TargetPickerProps = {
  targetPath: string
  onTargetPathChange: (targetPath: string) => void
}

export function TargetPicker({
  targetPath,
  onTargetPathChange,
}: TargetPickerProps) {
  return <div className='grow-1'>TargetPicker</div>
}
