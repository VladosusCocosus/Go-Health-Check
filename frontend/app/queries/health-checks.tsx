import {useQuery} from "@tanstack/react-query";
import {client} from "~/libs/axios";

export interface Config  {
    http: {
        name: string
        host: string
        endpoints: {
            path: string
            method: string
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
