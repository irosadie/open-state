"use client"

import type { FC } from "react"
import "filepond/dist/filepond.min.css"
import "filepond-plugin-image-preview/dist/filepond-plugin-image-preview.css"
import type { FilePondFile } from "filepond"
import FilePondPluginImagePreview from "filepond-plugin-image-preview"
import { FilePond, type FilePondProps, registerPlugin } from "react-filepond"
import { label as labelInput } from "./file.style"

registerPlugin(FilePondPluginImagePreview)

export type FileProps = FilePondProps & {
  label?: string
  error?: string
  hint?: string
  labelColor?: "main" | "primary" | "danger" | "success" | "warning" | "info"
  required?: boolean
}

const File: FC<FileProps> = (props) => {
  const { label, labelColor, required, hint, error, ...rest } = props

  return (
    <div className="space-y-1.5">
      {label && (
        <span className={labelInput({ intent: labelColor })}>
          {label}
          {required && <span className="text-danger-500">*</span>}
        </span>
      )}
      <div className="relative ">
        <FilePond allowMultiple={false} {...rest} />
        <div className="absolute -bottom-5 left-0">
          {hint && !error && (
            <span className="text-xs text-main-300">{hint}</span>
          )}
          {error && <span className="text-xs text-danger-500">{error}</span>}
        </div>
      </div>
    </div>
  )
}

export type { FilePondFile }
export default File
