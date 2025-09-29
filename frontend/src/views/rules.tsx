type RulesProps = {
  rules?: { prefix: string; eventName: string }
  setRules: (rules: { prefix: string; eventName: string }) => void
}

export function Rules({ rules, setRules }: RulesProps) {
  return <div className="grow-3">Rules</div>
}
