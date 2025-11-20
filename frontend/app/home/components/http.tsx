import type {Config, Status} from "~/queries/health-checks";
import cx from 'classnames'

interface HTTPsProps {
    httpConfig:  Config['http'] | undefined
    statuses?: Status[]
}

export function HTTPs ({httpConfig, statuses}: HTTPsProps) {
    if (!httpConfig) {
        return null
    }

    function findStatus (path: string) {
        return statuses?.find((s) => s.path === path)?.success
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
                                <span className={cx('w-2 h-2 rounded-full', findStatus(endpoints.path) ? 'bg-green-900 ': 'bg-red-900')}></span>
                                <span>{endpoints.path} ({endpoints.method})</span>
                            </div>
                        ))}
                    </div>
                </div>
            ))}
        </div>
    )
}
