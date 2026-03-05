import { z } from 'zod'

// Auth & User Schemas
export const authConfigSchema = z.object({
  serverName: z.string().min(1, 'Server name is required'),
  authType: z.enum(['oauth2', 'apikey', 'none']),
  clientId: z.string().optional(),
  clientSecret: z.string().optional(),
  authServerUrl: z.string().url().optional(),
  scopes: z.array(z.string()).default([]),
})
export type AuthConfig = z.infer<typeof authConfigSchema>

export const serverFormSchema = z.object({
  name: z.string().min(1, 'Server name is required').max(100),
  description: z.string().max(500).default(''),
  category: z.enum(['data', 'automation', 'ai', 'integration', 'other']),
  version: z.string().regex(/^\d+\.\d+\.\d+$/, 'Version must be semantic'),
  dockerImage: z.string().url('Must be a valid image URL'),
  author: z.string().min(1, 'Author is required'),
  homepage: z.string().url().optional(),
  license: z.string().default('MIT'),
})
export type ServerForm = z.infer<typeof serverFormSchema>

export const pricingModelSchema = z.object({
  type: z.enum(['free', 'subscription', 'flat', 'x402']),
  displayName: z.string().min(1),
  description: z.string().optional(),
  basePrice: z.number().positive().optional(),
  currency: z.enum(['USD', 'EUR', 'GBP']).default('USD'),
  perCallEnabled: z.boolean().default(false),
  perCallAmount: z.number().positive().optional(),
  caip2: z.string().optional(), // CAIP-2 asset identifier
  paymentAddress: z.string().optional(),
})
export type PricingModel = z.infer<typeof pricingModelSchema>

export const oauthConfigSchema = z.object({
  authServerUrl: z.string().url('Must be a valid URL'),
  clientId: z.string().min(1),
  clientSecret: z.string().min(1),
  metadataUrl: z.string().url().optional(),
  resourceUri: z.string().optional(),
  scopes: z.array(z.string()).default([]),
  registrationMode: z.enum(['pre-registered', 'cimd', 'dcr']).default('pre-registered'),
})
export type OAuthConfig = z.infer<typeof oauthConfigSchema>

export const deploymentSchema = z.object({
  environment: z.enum(['development', 'staging', 'production']),
  region: z.string().min(1),
  replicas: z.number().int().positive().default(1),
  transportType: z.enum(['http', 'grpc', 'websocket']).default('http'),
})
export type Deployment = z.infer<typeof deploymentSchema>

export const toolDefinitionSchema = z.object({
  name: z.string().min(1),
  description: z.string(),
  inputSchema: z.record(z.any()).optional(),
  requiredScopes: z.array(z.string()).default([]),
  enabled: z.boolean().default(true),
})
export type ToolDefinition = z.infer<typeof toolDefinitionSchema>
