import {useQuery} from "@tanstack/react-query";
import {client} from "~/libs/axios";

export interface Config  {
    http: {
        name: string
        host: string
        endpoints: {
            path: string
            method: string
            expectedStatus: number
        }[]
    }[]
    sftp: null
}

export function useHealthChecks () {
    return useQuery({
        queryKey: ['health-checks'],
        async queryFn () {
           const response = await client.get<Config>('config')

            return response.data
        }
    })
}

export interface Status  {
    success: boolean
    domain: string
    path: string
}

export function useHealthChecksStatuses () {
    return useQuery({
        queryKey: ['statuses'],
        async queryFn () {
            const response = await client.get<Status[]>('statuses')

            return response.data
        }
    })
}
