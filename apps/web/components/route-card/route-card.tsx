"use client"
import { ActionsDropdown } from "$/components/actions-dropdown"
import { PanelCard } from "$/components/panel-card"
import { cn } from "$/utils/cn"
import {
  ChevronDown,
  ChevronUp,
  Clock,
  MapPin,
  Navigation,
  Ship,
  Truck,
} from "lucide-react"

export type LocationType =
  | "WAREHOUSE"
  | "PORT"
  | "DISTRIBUTION_CENTER"
  | "CUSTOM_LOCATION"
export type TransportMode = "TRUCK" | "SHIP" | "PLANE" | "RAIL"

export type Route = {
  id: string
  name: string
  description?: string
  originLocationId: string
  originLocationName: string
  originType: LocationType
  destinationLocationId: string
  destinationLocationName: string
  destinationType: LocationType
  distance: number
  estimatedDuration: number
  stops: RouteStop[]
  isActive: boolean
  transportMode: "TRUCK" | "SHIP" | "PLANE" | "RAIL"
  createdAt: string
  updatedAt: string
}

export type RouteStop = {
  id: string
  locationId: string
  locationName: string
  locationType: LocationType
  address: string
  sequence: number
}

export const getLocationIcon = (type: LocationType, className = "w-4 h-4") => {
  switch (type) {
    case "WAREHOUSE":
      return <Navigation className={className} />

    case "PORT":
      return <Ship className={className} />

    case "DISTRIBUTION_CENTER":
      return <MapPin className={className} />

    default:
      return <MapPin className={className} />
  }
}

export const getLocationColor = (type: LocationType) => {
  switch (type) {
    case "WAREHOUSE":
      return "bg-orange-50 text-orange-600 border-orange-100"

    case "PORT":
      return "bg-blue-50 text-blue-600 border-blue-100"

    case "DISTRIBUTION_CENTER":
      return "bg-indigo-50 text-indigo-600 border-indigo-100"

    default:
      return "bg-gray-50 text-gray-600 border-gray-100"
  }
}

export const getTransportIcon = (mode: TransportMode) => {
  switch (mode) {
    case "SHIP":
      return <Ship className="w-6 h-6" />

    case "PLANE":
      return <Truck className="w-6 h-6" /> // Replace with plane icon

    case "RAIL":
      return <Truck className="w-6 h-6" /> // Replace with rail icon

    default:
      return <Truck className="w-6 h-6" />
  }
}

export const formatDuration = (minutes: number) => {
  if (!minutes) {
    return "-"
  }
  const hours = Math.floor(minutes / 60)
  const mins = minutes % 60

  if (hours >= 24) {
    const days = Math.floor(hours / 24)
    const remainingHours = hours % 24

    return `${days}d ${remainingHours}h`
  }
  return `${hours}h ${mins > 0 ? ` ${mins}m` : ""}`
}

export const RouteSummary = ({ route }: { route: Route }) => {
  const stopDots = Array.from(
    { length: Math.min(route.stops.length, 3) },
    (_, index) => ({
      id: `stop-dot-${index + 1}`,
      left: `${25 + index * 20}%`,
    }),
  )

  return (
    <div className="flex items-center justify-between bg-gray-50/80 rounded-xl p-4 border border-gray-100">
      <div className="flex items-center gap-3">
        <div
          className={cn(
            "p-2 rounded-lg bg-white border shadow-xs",
            getLocationColor(route.originType),
          )}
        >
          {getLocationIcon(route.originType)}
        </div>
        <div>
          <p className="text-xs font-semibold text-gray-500 uppercase">
            Origin
          </p>
          <p className="text-sm font-medium text-gray-900">
            {route.originLocationName}
          </p>
        </div>
      </div>

      {/* Connector Line */}
      <div className="flex-1 mx-6 flex flex-col items-center relative">
        {route.stops.length > 0 ? (
          <>
            <div className="w-full h-0.5 bg-gray-300 relative">
              <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 bg-gray-50 px-2 text-xs font-medium text-gray-500 flex items-center gap-1">
                {route.stops.length} stops
              </div>
              {stopDots.map((stopDot) => (
                <div
                  key={stopDot.id}
                  className="absolute top-1/2 -translate-y-1/2 w-1.5 h-1.5 rounded-full bg-gray-400 border border-white"
                  style={{ left: stopDot.left }}
                />
              ))}
            </div>
          </>
        ) : (
          <div className="w-full h-0.5 bg-gray-200" />
        )}
        <div className="absolute right-0 top-1/2 -translate-y-1/2 text-gray-300">
          <ArrowRight className="w-4 h-4" />
        </div>
      </div>

      <div className="flex items-center gap-3 text-right">
        <div>
          <p className="text-xs font-semibold text-gray-500 uppercase">
            Destination
          </p>
          <p className="text-sm font-medium text-gray-900">
            {route.destinationLocationName}
          </p>
        </div>
        <div
          className={cn(
            "p-2 rounded-lg bg-white border shadow-sm",
            getLocationColor(route.destinationType),
          )}
        >
          {getLocationIcon(route.destinationType)}
        </div>
      </div>
    </div>
  )
}

// --- Route Timeline Component ---

type TimelineItem = {
  id: string
  name: string
  type: LocationType
  address?: string
  isStart?: boolean
  isEnd?: boolean
  isStop?: boolean
}

export const RouteTimeline = ({ route }: { route: Route }) => {
  const timelineItems: TimelineItem[] = [
    {
      id: "origin",
      name: route.originLocationName,
      type: route.originType,
      address: "Origin",
      isStart: true,
    },
    ...route.stops
      .sort((a, b) => a.sequence - b.sequence)
      .map((stop) => ({
        id: stop.id,
        name: stop.locationName,
        type: stop.locationType as LocationType,
        address: stop.address,
        isStop: true,
      })),
    {
      id: "destination",
      name: route.destinationLocationName,
      type: route.destinationType,
      address: "Destination",
      isEnd: true,
    },
  ]

  return (
    <div className="relative pl-4 space-y-0 text-sm">
      {/* Vertical Line */}
      <div className="absolute top-4 bottom-4 left-5.25 w-0.5 bg-gray-200" />

      {timelineItems.map((item) => (
        <div
          key={item.id}
          className="relative flex items-start gap-4 pb-6 last:pb-0 group"
        >
          {/* Dot / Icon */}
          <div
            className={cn(
              "relative z-10 flex items-center justify-center w-8 h-8 rounded-full border-2 bg-white transition-colors",
              item.isStart
                ? "border-green-500 text-green-600"
                : item.isEnd
                  ? "border-red-500 text-red-600"
                  : "border-gray-300 text-gray-500 group-hover:border-primary-400 group-hover:text-primary-500",
            )}
          >
            {item.isStart || item.isEnd ? (
              <div
                className={cn(
                  "w-2.5 h-2.5 rounded-full",
                  item.isStart ? "bg-green-500" : "bg-red-500",
                )}
              />
            ) : (
              <div className="w-1.5 h-1.5 rounded-full bg-gray-300 group-hover:bg-primary-500" />
            )}
          </div>

          {/* Content */}
          <div className="flex-1 pt-1">
            <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-1 mb-1">
              <span
                className={cn(
                  "font-medium",
                  item.isStart || item.isEnd
                    ? "text-gray-900"
                    : "text-gray-700",
                )}
              >
                {item.name}
              </span>
              <span
                className={cn(
                  "text-[10px] px-2 py-0.5 rounded-full border uppercase tracking-wider font-semibold w-fit",
                  getLocationColor(item.type),
                )}
              >
                {item.type.replace("_", " ")}
              </span>
            </div>
            {item.address && (
              <p className="text-xs text-gray-500">{item.address}</p>
            )}
          </div>
        </div>
      ))}
    </div>
  )
}

interface RouteCardProps {
  route: Route
  isExpanded: boolean
  onToggleExpand: () => void
  onEdit?: () => void
  onDelete?: () => void
}

export const RouteCard = ({
  route,
  isExpanded,
  onToggleExpand,
  onEdit,
  onDelete,
}: RouteCardProps) => {
  return (
    <PanelCard
      className={cn(
        "group transition-all duration-300",
        isExpanded && "ring-1 ring-primary-100",
      )}
    >
      <div className="p-0">
        {/* Card Header */}
        <div className="flex items-start justify-between mb-6">
          <div className="flex items-center gap-4">
            <div
              className={cn(
                "w-12 h-12 rounded-xl flex items-center justify-center transition-colors",
                route.isActive
                  ? "bg-primary-50 text-primary-600"
                  : "bg-gray-100 text-gray-400",
              )}
            >
              {getTransportIcon(route.transportMode)}
            </div>
            <div>
              <div className="flex items-center gap-2">
                <h3 className="text-lg font-bold text-gray-900 group-hover:text-primary-700 transition-colors">
                  {route.name}
                </h3>
                {!route.isActive && (
                  <span className="px-2 py-0.5 rounded text-[10px] font-bold bg-gray-100 text-gray-500 uppercase tracking-wide">
                    Inactive
                  </span>
                )}
              </div>
              {route.description && (
                <p className="mt-1 text-xs text-gray-500">
                  {route.description}
                </p>
              )}
              <div className="flex items-center gap-3 mt-1 text-xs text-gray-500">
                <span className="flex items-center gap-1">
                  <MapPin className="w-3 h-3" />{" "}
                  {route.distance.toLocaleString()} km
                </span>
                <span className="w-1 h-1 rounded-full bg-gray-300" />
                <span className="flex items-center gap-1">
                  <Clock className="w-3 h-3" />{" "}
                  {formatDuration(route.estimatedDuration)}
                </span>
              </div>
            </div>
          </div>

          {/* Actions */}
          <div className="opacity-0 group-hover:opacity-100 transition-opacity">
            <ActionsDropdown
              actions={[
                {
                  label: "Edit",
                  onClick: onEdit || (() => {}),
                },
                {
                  label: "Delete",
                  onClick: onDelete || (() => {}),
                  destructive: true,
                },
              ]}
            />
          </div>
        </div>

        {/* Visual Route Representation (Horizontal Summary) */}
        {!isExpanded && <RouteSummary route={route} />}

        {/* Expanded View (Detailed Timeline) */}
        {isExpanded && (
          <div className="mt-6 animate-in slide-in-from-top-2 duration-300">
            <div className="bg-gray-50 rounded-xl p-6 border border-gray-100">
              <h4 className="text-xs font-bold text-gray-400 uppercase tracking-widest mb-6">
                Full Route Path
              </h4>
              <RouteTimeline route={route} />
            </div>
          </div>
        )}

        {/* Toggle Button */}
        <div className="mt-4 flex justify-center">
          <button
            type="button"
            onClick={onToggleExpand}
            className="flex items-center gap-1 text-xs font-medium text-gray-400 hover:text-primary-600 transition-colors py-1 px-3 rounded-full hover:bg-primary-50"
          >
            {isExpanded ? (
              <>
                Collapse <ChevronUp className="w-3 h-3" />
              </>
            ) : (
              <>
                View Details <ChevronDown className="w-3 h-3" />
              </>
            )}
          </button>
        </div>
      </div>
    </PanelCard>
  )
}

// --- ArrowRight Icon for RouteSummary ---
const ArrowRight = ({ className }: { className?: string }) => (
  <svg
    aria-hidden="true"
    className={className}
    fill="none"
    viewBox="0 0 24 24"
    stroke="currentColor"
  >
    <title>Arrow right</title>
    <path
      strokeLinecap="round"
      strokeLinejoin="round"
      strokeWidth={2}
      d="M17 8l4 4m0 0l-4 4m4-4H3"
    />
  </svg>
)
