import { Link } from '@tanstack/react-router'
import { ChevronDown, Menu } from 'lucide-react'
import * as React from 'react'
import { useMemo } from 'react'

import { Button } from '@/components/ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { cn } from '@/lib/utils'

import type { TopNavLink } from '../types'

type TopNavProps = React.HTMLAttributes<HTMLElement> & {
  links: TopNavLink[]
}

/**
 * 顶部导航栏组件
 * 在大屏幕显示水平导航，在小屏幕显示下拉菜单
 */
export function TopNav({ className, links, ...props }: TopNavProps) {
  // 规范化链接，确保所有可选属性都有默认值
  const normalizedLinks = useMemo(
    () =>
      links.map((link) => ({
        isActive: false,
        disabled: false,
        external: false,
        ...link,
      })),
    [links]
  )

  return (
    <>
      {/* 移动端下拉菜单 */}
      <div className='lg:hidden'>
        <DropdownMenu modal={false}>
          <DropdownMenuTrigger
            render={<Button size='icon' variant='outline' className='size-7' />}
          >
            <Menu />
          </DropdownMenuTrigger>
          <DropdownMenuContent side='bottom' align='start'>
            {normalizedLinks.map((link) =>
              link.dropdownItems ? (
                <React.Fragment key={link.title}>
                  <DropdownMenuItem disabled>{link.title}</DropdownMenuItem>
                  {link.dropdownItems.map((item) => (
                    <DropdownMenuItem
                      key={`${link.title}-${item.title}`}
                      className='pl-5'
                      render={
                        item.external ? (
                          <a
                            href={item.href}
                            target='_blank'
                            rel='noopener noreferrer'
                          >
                            {item.title}
                          </a>
                        ) : (
                          <Link to={item.href} disabled={item.disabled}>
                            {item.title}
                          </Link>
                        )
                      }
                    />
                  ))}
                </React.Fragment>
              ) : (
                <DropdownMenuItem
                  key={`${link.title}-${link.href}`}
                  render={
                    link.external ? (
                      <a
                        href={link.href}
                        target='_blank'
                        rel='noopener noreferrer'
                      >
                        {link.title}
                      </a>
                    ) : (
                      <Link to={link.href} disabled={link.disabled}>
                        {link.title}
                      </Link>
                    )
                  }
                />
              )
            )}
          </DropdownMenuContent>
        </DropdownMenu>
      </div>

      {/* 桌面端水平导航 */}
      <nav
        className={cn(
          'hidden items-center space-x-4 lg:flex lg:space-x-4 xl:space-x-6',
          className
        )}
        {...props}
      >
        {normalizedLinks.map((link) =>
          // oxlint-disable-next-line no-nested-ternary
          link.dropdownItems ? (
            <DropdownMenu key={link.title} modal={false}>
              <DropdownMenuTrigger className='text-muted-foreground hover:text-foreground inline-flex items-center gap-1 text-sm font-medium transition-colors'>
                {link.title}
                <ChevronDown className='size-3.5' aria-hidden='true' />
              </DropdownMenuTrigger>
              <DropdownMenuContent align='start' className='min-w-40 p-2'>
                {link.dropdownItems.map((item) => (
                  <DropdownMenuItem
                    key={`${link.title}-${item.title}`}
                    render={
                      item.external ? (
                        <a
                          href={item.href}
                          target='_blank'
                          rel='noopener noreferrer'
                        >
                          {item.title}
                        </a>
                      ) : (
                        <Link to={item.href} disabled={item.disabled}>
                          {item.title}
                        </Link>
                      )
                    }
                  />
                ))}
              </DropdownMenuContent>
            </DropdownMenu>
          ) : link.external ? (
            <a
              key={`${link.title}-${link.href}`}
              href={link.href}
              target='_blank'
              rel='noopener noreferrer'
              className={`hover:text-primary text-sm font-medium transition-colors ${link.isActive ? '' : 'text-muted-foreground'}`}
            >
              {link.title}
            </a>
          ) : (
            <Link
              key={`${link.title}-${link.href}`}
              to={link.href}
              disabled={link.disabled}
              className={`hover:text-primary text-sm font-medium transition-colors ${link.isActive ? '' : 'text-muted-foreground'}`}
            >
              {link.title}
            </Link>
          )
        )}
      </nav>
    </>
  )
}
