import { Button } from "@/components/ui/button"

type ExcuteProps = { onSubmit: () => void }

export function Excute({ onSubmit }: ExcuteProps) {
  return (
    <div className='flex items-center justify-end-safe gap-2 grow-0'>
      <Button variant='destructive'>Reset</Button>
      <Button disabled>Proceed</Button>
    </div>
  )
}
