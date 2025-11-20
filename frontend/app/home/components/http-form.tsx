import {useState} from "react";
import {useFieldArray, useForm} from "react-hook-form";
import {z} from "zod";
import {zodResolver} from "@hookform/resolvers/zod";
import {useMutation, useQueryClient} from "@tanstack/react-query";

import {client} from "~/libs/axios";

const httpMethods = ["GET", "POST", "PUT", "PATCH", "DELETE", "HEAD"] as const

const endpointSchema = z.object({
    path: z.string().min(1, "Path is required"),
    method: z.enum(httpMethods),
    expectedStatus: z.coerce.number().int().min(100).max(599).default(200),
    schedule: z.string().min(5, "Schedule is requred")
})

const httpSchema = z.object({
    name: z.string().min(1, "Name is required"),
    host: z.string(),
    endpoints: z.array(endpointSchema).min(1, "Add at least one endpoint"),
})

type HttpFormValues = z.infer<typeof httpSchema>

const defaultValues: HttpFormValues = {
    name: "",
    host: "",
    endpoints: [
        {
            path: "",
            method: "GET",
            expectedStatus: 200,
            schedule: "*/5 * * * *"
        },
    ],
}

export function HttpForm() {
    const queryClient = useQueryClient()
    const [submissionError, setSubmissionError] = useState<string | null>(null)

    const {
        register,
        control,
        handleSubmit,
        reset,
        formState: {errors},
    } = useForm({
        resolver: zodResolver(httpSchema),
        defaultValues,
    })

    const {fields, append, remove} = useFieldArray({
        control,
        name: "endpoints",
    })

    const mutation = useMutation({
        mutationFn: async (values: HttpFormValues) => {
            await client.post("/config/http", values)
        },
        onSuccess: () => {
            queryClient.invalidateQueries({queryKey: ["health-checks"]})
            reset(defaultValues)
        },
        onError: (error: unknown) => {
            const message = error instanceof Error ? error.message : "Failed to save configuration"
            setSubmissionError(message)
        },
    })

    const onSubmit = (values: HttpFormValues) => {
        setSubmissionError(null)
        mutation.mutate(values)
    }

    return (
        <form onSubmit={handleSubmit(onSubmit)} className={'flex flex-col gap-6 border border-slate-200 rounded-lg p-6 bg-white shadow-sm max-w-3xl'}>
            <div className={'flex flex-col gap-2'}>
                <label className={'text-sm font-medium text-slate-900'} htmlFor={'http-name'}>Service name</label>
                <input
                    id={'http-name'}
                    type={'text'}
                    placeholder={'e.g. Cat Facts'}
                    className={'border border-slate-200 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-blue-400'}
                    {...register('name')}
                />
                {errors.name && <p className={'text-sm text-red-600'}>{errors.name.message}</p>}
            </div>

            <div className={'flex flex-col gap-2'}>
                <label className={'text-sm font-medium text-slate-900'} htmlFor={'http-host'}>Base URL</label>
                <input
                    id={'http-host'}
                    type={'text'}
                    placeholder={'https://api.example.com'}
                    className={'border border-slate-200 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-blue-400'}
                    {...register('host')}
                />
                {errors.host && <p className={'text-sm text-red-600'}>{errors.host.message}</p>}
            </div>

            <div className={'flex flex-col gap-3'}>
                <div className={'flex items-center justify-between'}>
                    <div>
                        <p className={'text-sm font-medium text-slate-900'}>Endpoints</p>
                        <p className={'text-sm text-slate-500'}>Specify each path, HTTP method, and expected status.</p>
                    </div>
                    <button
                        type={'button'}
                        className={'text-sm text-blue-600 hover:text-blue-700 font-medium'}
                        onClick={() => append({path: "", method: "GET", expectedStatus: 200, schedule: "*/5 * * * *"})}
                    >
                        + Add endpoint
                    </button>
                </div>

                {fields.map((field, index) => (
                    <div key={field.id} className={'grid grid-cols-1 md:grid-cols-[1fr_160px_160px] gap-3 items-start border border-slate-200 rounded-md p-4'}>
                        <div className={'flex flex-col gap-2'}>
                            <label className={'text-sm font-medium text-slate-900'} htmlFor={`endpoint-path-${index}`}>
                                Path
                            </label>
                            <input
                                id={`endpoint-path-${index}`}
                                type={'text'}
                                placeholder={'/health'}
                                className={'border border-slate-200 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-blue-400'}
                                {...register(`endpoints.${index}.path` as const)}
                            />
                            {errors.endpoints?.[index]?.path && (
                                <p className={'text-sm text-red-600'}>{errors.endpoints[index]?.path?.message}</p>
                            )}
                        </div>

                        <div className={'flex flex-col gap-2'}>
                            <label className={'text-sm font-medium text-slate-900'} htmlFor={`endpoint-method-${index}`}>
                                Method
                            </label>
                            <select
                                id={`endpoint-method-${index}`}
                                className={'border border-slate-200 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-blue-400 bg-white'}
                                {...register(`endpoints.${index}.method` as const)}
                            >
                                {httpMethods.map((method) => (
                                    <option value={method} key={method}>{method}</option>
                                ))}
                            </select>
                            {errors.endpoints?.[index]?.method && (
                                <p className={'text-sm text-red-600'}>{errors.endpoints[index]?.method?.message}</p>
                            )}
                        </div>

                        <div className={'flex flex-col gap-2'}>
                            <label className={'text-sm font-medium text-slate-900'} htmlFor={`endpoint-expected-${index}`}>
                                Expected status
                            </label>
                            <input
                                id={`endpoint-expected-${index}`}
                                type={'number'}
                                min={100}
                                max={599}
                                className={'border border-slate-200 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-blue-400'}
                                {...register(`endpoints.${index}.expectedStatus` as const, {valueAsNumber: true})}
                            />
                            {errors.endpoints?.[index]?.expectedStatus && (
                                <p className={'text-sm text-red-600'}>{errors.endpoints[index]?.expectedStatus?.message}</p>
                            )}
                        </div>

                        <div className={'flex flex-col gap-2'}>
                            <label className={'text-sm font-medium text-slate-900'} htmlFor={`endpoint-expected-${index}`}>
                                Schedule (Cron format)
                            </label>
                            <input
                                id={`endpoint-schedule-cron-${index}`}
                                type={'text'}
                                className={'border border-slate-200 rounded-md px-3 py-2 focus:outline-none focus:ring-2 focus:ring-blue-400 focus:border-blue-400'}
                                {...register(`endpoints.${index}.schedule` as const)}
                            />
                            {errors.endpoints?.[index]?.schedule && (
                                <p className={'text-sm text-red-600'}>{errors.endpoints[index]?.schedule?.message}</p>
                            )}
                        </div>

                        {fields.length > 1 && (
                            <div className={'md:col-span-3 flex justify-end'}>
                                <button
                                    type={'button'}
                                    className={'text-sm text-red-600 hover:text-red-700 font-medium'}
                                    onClick={() => remove(index)}
                                >
                                    Remove
                                </button>
                            </div>
                        )}
                    </div>
                ))}

            </div>

            {submissionError && <p className={'text-sm text-red-600'}>{submissionError}</p>}

            <div className={'flex justify-end'}>
                <button
                    type={'submit'}
                    disabled={mutation.isPending}
                    className={'bg-blue-600 text-white px-4 py-2 rounded-md text-sm font-medium hover:bg-blue-700 disabled:opacity-60 disabled:cursor-not-allowed'}
                >
                    {mutation.isPending ? 'Saving...' : 'Save HTTP config'}
                </button>
            </div>
        </form>
    )
}
