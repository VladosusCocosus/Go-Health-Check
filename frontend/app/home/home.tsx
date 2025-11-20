import {useHealthChecks, useHealthChecksStatuses} from "~/queries/health-checks";
import {HTTPs} from "~/home/components/http";
import {useState} from "react";
import {HttpForm} from "~/home/components/http-form";

export function Home() {
    const {data} = useHealthChecks()
    const {data: dataStatuses} = useHealthChecksStatuses()
    const [newHttp, setNewHttp] = useState(false)

  return (
    <div className={'p-10 flex flex-col gap-10'}>
        <h1 className={'text-xl text-zinc-900 font-bold tracking-wide'}>Health Check Application</h1>
        <HTTPs httpConfig={data?.http} statuses={dataStatuses}/>
        <button onClick={() => setNewHttp((prev) => !prev)}>Add new</button>
        {newHttp && (
            <HttpForm/>
        )}
    </div>
  );
}




