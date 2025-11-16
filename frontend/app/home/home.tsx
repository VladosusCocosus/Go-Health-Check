import {type Config, useHealthChecks} from "~/queries/health-checks";

export function Home() {
    const {data} = useHealthChecks()

  return (
    <div className={'p-10 flex flex-col gap-10'}>
        <h1 className={'text-xl text-zinc-900 font-bold tracking-wide'}>Health Check Application</h1>
        <HTTPs httpConfig={data?.http}/>
    </div>
  );
}


interface HTTPsProps {
    httpConfig:  Config['http'] | undefined
}

function HTTPs ({httpConfig}: HTTPsProps) {
    if (!httpConfig) {
        return null
    }

    return (
        <div className={'flex flex-row flex-wrap gap-10'}>
            {httpConfig.map((http) => (
                <div key={http.host} className={'flex flex-col gap-4'}>
                    <a
                        className={'text-xl text-zinc-900 border-b-2 border-transparent hover:border-gray-400 '}
                        href={http.host}
                        target={'_blank'}
                    >
                        {http.name}
                    </a>

                    <div>
                    {http.endpoints.map((endpoints) => (
                        <div key={endpoints.path} className={'flex flex-row gap-4 items-center'}>
                            <span className={'w-2 h-2 bg-red-900 rounded-full'}></span>
                            <span>{endpoints.path} ({endpoints.method})</span>
                        </div>
                    ))}
                    </div>
                </div>
            ))}
        </div>
    )
}

