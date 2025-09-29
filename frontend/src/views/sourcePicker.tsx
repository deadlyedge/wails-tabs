type SourcePickerProps = {
  sourcePath: string
  onSourcePathChange: (sourcePath: string) => void
}

export default function SourcePicker({
  sourcePath,
  onSourcePathChange,
}: SourcePickerProps) {
  return <div className="grow-1">SourcePicker</div>
}
