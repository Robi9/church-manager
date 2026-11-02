"use client";

import { useState, type FormEvent } from "react";
import Image from "next/image";
import { useAuth } from "@/lib/auth";
import { api } from "@/lib/api";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Loader2, ShieldCheck } from "lucide-react";

export default function LoginPage() {
  const { login } = useAuth();
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  async function handleSubmit(e: FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);

    try {
      const res = await api<{ token: string }>("/auth/login", {
        method: "POST",
        body: { email, password },
      });
      login(res.data.token);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Login failed");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="brand-surface relative flex min-h-screen items-center justify-center overflow-hidden px-4 py-10">
      <div className="pointer-events-none absolute -left-24 -top-24 h-72 w-72 rounded-full border-[42px] border-primary/5" />
      <div className="pointer-events-none absolute -bottom-32 -right-24 h-96 w-96 rounded-full border-[58px] border-secondary/20" />

      <div className="relative grid w-full max-w-5xl overflow-hidden rounded-3xl border border-white/70 bg-white/90 shadow-[0_30px_90px_rgba(102,48,38,0.16)] backdrop-blur lg:grid-cols-[1.08fr_0.92fr]">
        <section className="relative hidden min-h-[620px] flex-col justify-between overflow-hidden border-r border-border bg-[#fff7e9] p-12 lg:flex">
          <div className="absolute -right-24 -top-20 h-72 w-72 rounded-full bg-accent/35 blur-2xl" />
          <div className="relative h-64 w-full">
            <Image
              src="/brand/logo-colorida.png"
              alt="Igreja Betesda Central"
              fill
              priority
              sizes="500px"
              className="object-contain"
            />
          </div>
          <div className="relative space-y-4">
            <p className="text-sm font-semibold uppercase tracking-[0.28em] text-primary">
              Ao Senhor · Aos santos · Ao mundo
            </p>
            <h1 className="max-w-md text-4xl font-bold leading-tight text-foreground">
              Cuidando de pessoas com organização e propósito.
            </h1>
            <p className="max-w-md text-sm leading-7 text-muted-foreground">
              Um ambiente seguro para apoiar o cuidado, a comunhão e a missão da Igreja Betesda Central.
            </p>
          </div>
          <div className="relative flex items-center gap-2 text-xs font-medium text-muted-foreground">
            <ShieldCheck className="h-4 w-4 text-primary" />
            Acesso exclusivo para a equipe autorizada
          </div>
        </section>

        <Card className="w-full justify-center rounded-none border-0 bg-transparent px-3 py-8 shadow-none ring-0 sm:px-10 lg:px-12">
          <CardHeader className="text-center lg:text-left">
            <div className="relative mx-auto mb-3 h-32 w-full max-w-xs lg:hidden">
              <Image
                src="/brand/logo-colorida.png"
                alt="Igreja Betesda Central"
                fill
                priority
                sizes="320px"
                className="object-contain"
              />
            </div>
            <p className="text-xs font-semibold uppercase tracking-[0.2em] text-primary">Área administrativa</p>
            <CardTitle className="text-3xl font-bold">Bem-vindo</CardTitle>
            <CardDescription>Entre com seus dados para acessar a gestão de membros.</CardDescription>
          </CardHeader>
          <CardContent className="pt-3">
          <form onSubmit={handleSubmit} className="space-y-4">
            {error && (
              <div className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
                {error}
              </div>
            )}
            <div className="space-y-2">
              <Label htmlFor="email">Email</Label>
              <Input
                id="email"
                type="email"
                placeholder="seu@email.com"
                autoComplete="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
                required
              />
            </div>
            <div className="space-y-2">
              <Label htmlFor="password">Senha</Label>
              <Input
                id="password"
                type="password"
                placeholder="Sua senha"
                autoComplete="current-password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                required
              />
            </div>
            <Button type="submit" className="h-11 w-full shadow-[0_8px_24px_rgba(171,51,53,0.22)]" disabled={loading}>
              {loading ? (
                <Loader2 className="h-4 w-4 animate-spin" />
              ) : (
                "Entrar"
              )}
            </Button>
          </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
