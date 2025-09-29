import { useState } from "react"
import "./App.css"
import SourcePicker from "./views/sourcePicker"
import TargetPicker from "./views/targetPicker"
import Rules from "./views/rules"
import Excute from "./views/excute"
import { Monitor } from "./views/monitor"
import { Separator } from "./components/ui/separator"

function App() {
  const onSourcePathChange = (sourcePath: string) => {
    console.log("Source path changed to", sourcePath)
  }

  return (
    <main className='h-svh w-full overflow-hidden'>
      <h1 className="text-2xl m-2 font-bold">photoTidyGo</h1>
      <div className='w-full h-[550px] flex'>
        <div className='w-1/2 flex flex-col p-2 gap-2 justify-stretch'>
          <SourcePicker
            sourcePath='/'
            onSourcePathChange={onSourcePathChange}
          />
          <Separator className="m-2" />
          <TargetPicker targetPath='/' onTargetPathChange={() => {}} />
          <Separator className="m-2" />
          <Rules setRules={() => {}} />
          <Separator className="m-2" />
          <Excute onSubmit={() => console.log("Submit")} />
        </div>
        <div className='w-1/2 flex flex-col items-center justify-center'>
          <Monitor showing={""} />
        </div>
      </div>
    </main>
  )
}

export default App
