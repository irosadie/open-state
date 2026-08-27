"use client"

import L from "leaflet"
import { useEffect, useRef } from "react"
import "leaflet/dist/leaflet.css"

// Fix for default markers in Next.js
;(L.Icon.Default.prototype as unknown as Record<string, unknown>)._getIconUrl =
  undefined
L.Icon.Default.mergeOptions({
  iconRetinaUrl:
    "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon-2x.png",
  iconUrl:
    "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-icon.png",
  shadowUrl:
    "https://cdnjs.cloudflare.com/ajax/libs/leaflet/1.9.4/images/marker-shadow.png",
})

interface Route {
  id: string
  origin: [number, number]
  destination: [number, number]
  waypoints?: [number, number][] // Optional intermediate waypoints for more accurate routes
  pathCoordinates?: [number, number][] // Precomputed route polyline provided by the caller
  currentPosition?: [number, number] // Current vehicle position on the route
}

export interface MapProps {
  center?: [number, number]
  zoom?: number
  markers?: Array<{
    id: string
    position: [number, number]
    heading: number
    type: "truck" | "ship"
    status: "moving" | "stopped" | "delayed" | "offline"
    onClick?: () => void
  }>
  routes?: Route[]
  className?: string
}

// Custom vehicle icons
const createVehicleIcon = (
  type: "truck" | "ship",
  status: string,
  heading: number,
) => {
  const colors: Record<string, string> = {
    moving: "#22c55e",
    stopped: "#9ca3af",
    delayed: "#f59e0b",
    offline: "#ef4444",
  }

  const color = colors[status] || "#6b7280"

  const svg =
    type === "ship"
      ? `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="${color}" width="32" height="32" style="transform: rotate(${heading}deg)">
				<path d="M3 17h18v2H3v-2zm3-9l2-3h8l2 3 2 5H6l-2-2z"/>
				<path d="M12 2L4 8h16L12 2z" fill="${color}"/>
			</svg>`
      : `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="${color}" width="32" height="32" style="transform: rotate(${heading}deg)">
				<path d="M20 8h-3V4H3c-1.1 0-2 .9-2 2v11h2c0 1.66 1.34 3 3 3s3-1.34 3-3h6c0 1.66 1.34 3 3 3s3-1.34 3-3h2v-5l-3-4zM6 18.5c-.83 0-1.5-.67-1.5-1.5s.67-1.5 1.5-1.5 1.5.67 1.5 1.5-.67 1.5-1.5 1.5zm13.5-9l1.96 2.5H17V9.5h2.5zm-1.5 9c-.83 0-1.5-.67-1.5-1.5s.67-1.5 1.5-1.5 1.5.67 1.5 1.5-.67 1.5-1.5 1.5z"/>
			</svg>`

  return L.divIcon({
    html: svg,
    className: "vehicle-marker",
    iconSize: [32, 32],
    iconAnchor: [16, 16],
    popupAnchor: [0, -16],
  })
}

// Create origin/destination markers
const createPinIcon = (type: "origin" | "destination") => {
  const color = type === "origin" ? "#3b82f6" : "#ef4444"
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="${color}" width="24" height="24">
		<path d="M12 2C8.13 2 5 5.13 5 9c0 5.25 7 13 7 13s7-7.75 7-13c0-3.87-3.13-7-7-7zm0 9.5c-1.38 0-2.5-1.12-2.5-2.5s1.12-2.5 2.5-2.5 2.5 1.12 2.5 2.5-1.12 2.5-2.5 2.5z"/>
	</svg>`

  return L.divIcon({
    html: svg,
    className: "pin-marker",
    iconSize: [24, 24],
    iconAnchor: [12, 24],
    popupAnchor: [0, -24],
  })
}

export default function RouteMap({
  center = [-0.7893, 113.9213],
  zoom = 5,
  markers = [],
  routes = [],
  className = "",
}: MapProps) {
  const mapRef = useRef<HTMLDivElement>(null)
  const mapInstanceRef = useRef<L.Map | null>(null)
  const routeLayersRef = useRef<L.LayerGroup | null>(null)

  useEffect(() => {
    if (!mapRef.current || mapInstanceRef.current) {
      return
    }

    // Initialize map
    mapInstanceRef.current = L.map(mapRef.current, {
      center,
      zoom,
      zoomControl: false,
      attributionControl: false,
    })

    // Add OpenStreetMap tiles
    L.tileLayer("https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png", {
      maxZoom: 19,
    }).addTo(mapInstanceRef.current)

    // Create layer group for routes
    routeLayersRef.current = L.layerGroup().addTo(mapInstanceRef.current)

    return () => {
      mapInstanceRef.current?.remove()
      mapInstanceRef.current = null
    }
  }, [center, zoom])

  // Update markers
  useEffect(() => {
    const mapInstance = mapInstanceRef.current

    if (!mapInstance) {
      return
    }

    // Clear existing markers
    mapInstance.eachLayer((layer) => {
      if (layer instanceof L.Marker) {
        mapInstance.removeLayer(layer)
      }
    })

    // Add new markers
    for (const marker of markers) {
      const icon = createVehicleIcon(marker.type, marker.status, marker.heading)
      const leafMarker = L.marker(marker.position, { icon })

      leafMarker.on("click", () => marker.onClick?.())
      leafMarker.addTo(mapInstance)
    }
  }, [markers])

  // Draw routes
  useEffect(() => {
    const routeLayers = routeLayersRef.current

    if (!mapInstanceRef.current || !routeLayers) {
      return
    }

    // Clear existing routes
    routeLayers.clearLayers()

    // Draw each route
    for (const route of routes) {
      const routeGroup = L.featureGroup()

      // Add origin marker
      const originMarker = L.marker(route.origin, {
        icon: createPinIcon("origin"),
      }).bindPopup(`Origin: ${route.origin.join(", ")}`)

      routeGroup.addLayer(originMarker)

      // Add destination marker
      const destMarker = L.marker(route.destination, {
        icon: createPinIcon("destination"),
      }).bindPopup(`Destination: ${route.destination.join(", ")}`)

      routeGroup.addLayer(destMarker)

      const routeCoordinates =
        route.pathCoordinates && route.pathCoordinates.length > 1
          ? route.pathCoordinates
          : [route.origin, ...(route.waypoints ?? []), route.destination]

      if (route.currentPosition) {
        // Find closest point on route to current position
        let closestIndex = 0
        let minDistance = Number.POSITIVE_INFINITY

        for (const [index, coord] of routeCoordinates.entries()) {
          const distance = Math.sqrt(
            (coord[0] - route.currentPosition?.[0]) ** 2 +
              (coord[1] - route.currentPosition?.[1]) ** 2,
          )

          if (distance < minDistance) {
            minDistance = distance
            closestIndex = index
          }
        }

        // Split route into completed (blue) and remaining (red)
        const completedRoute = routeCoordinates.slice(0, closestIndex + 1)
        const remainingRoute = routeCoordinates.slice(closestIndex)

        if (completedRoute.length > 1) {
          const completedLine = L.polyline(completedRoute, {
            color: "#3b82f6",
            weight: 4,
            opacity: 0.9,
          })

          routeGroup.addLayer(completedLine)
        }

        if (remainingRoute.length > 1) {
          const remainingLine = L.polyline(remainingRoute, {
            color: "#ef4444",
            weight: 4,
            opacity: 0.8,
            dashArray: "10, 10",
          })

          routeGroup.addLayer(remainingLine)
        }
      } else {
        const routeLine = L.polyline(routeCoordinates, {
          color: "#ef4444",
          weight: 4,
          opacity: 0.8,
          dashArray: "10, 10",
        })

        routeGroup.addLayer(routeLine)
      }

      routeLayers.addLayer(routeGroup)
    }
  }, [routes])

  return <div ref={mapRef} className={className} />
}
